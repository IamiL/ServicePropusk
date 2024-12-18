package minioRepository

import (
	"context"
	"errors"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"log"
	"mime/multipart"
	"os"
	model "rip/internal/domain"
	postgresBuilds "rip/internal/repository/postgres/builds"
	"strings"
)

func Connect(
	endpoint string,
	accessKeyID string,
	secretAccessKey string,
) (*minio.Client, error) {
	useSSL := false

	// Initialize minio client object.
	minioClient, err := minio.New(
		endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
			Secure: useSSL,
		},
	)
	if err != nil {
		log.Fatalln(err)
	}

	buckets, err := minioClient.ListBuckets(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("connection to minio successful")

	for i, bucket := range buckets {
		log.Println("bucket № ", i, " - ", bucket)
	}

	return minioClient, nil
}

type MinioRepository struct {
	Session                   *minio.Client
	BuildingsPhotosBucketName string
	StaticFilesBucketName     string
	PhotosLocalPath           string
	StaticFilesPath           string
	buildingsRepo             *postgresBuilds.Storage
}

func New(
	sess *minio.Client,
	buildsPhotosBucketName string,
	staticFilesBucketName string,
	photosLocalPath string,
	staticFilesPath string,
	buildingsRepo *postgresBuilds.Storage,
) *MinioRepository {
	return &MinioRepository{
		sess,
		buildsPhotosBucketName,
		staticFilesBucketName,
		photosLocalPath,
		staticFilesPath,
		buildingsRepo,
	}
}

func (s *MinioRepository) ConfigureMinioStorage() error {
	found, err := s.Session.BucketExists(
		context.Background(),
		s.BuildingsPhotosBucketName,
	)
	if err != nil {
		log.Fatalln(err)
	}

	if found {
		log.Println("Bucket found.")
	} else {
		log.Println("Bucket not found.")

		log.Println("Creating minio bucket start")

		opts := minio.MakeBucketOptions{
			ObjectLocking: false,
			Region:        "us-east-1",
		}

		err = s.Session.MakeBucket(
			context.Background(),
			s.BuildingsPhotosBucketName,
			opts,
		)
		if err != nil {
			log.Fatalln("makebucket error - ", err.Error())
		}

		policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetBucketLocation","s3:ListBucket","s3:ListBucketMultipartUploads"],"Resource":["arn:aws:s3:::services"]},{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:ListMultipartUploadParts","s3:PutObject","s3:AbortMultipartUpload","s3:DeleteObject","s3:GetObject"],"Resource":["arn:aws:s3:::services/*"]}]}`

		err = s.Session.SetBucketPolicy(
			context.Background(),
			s.BuildingsPhotosBucketName,
			policy,
		)
		if err != nil {
			log.Fatalln("SetBucketPolicy error - ", err.Error())
		}

		if err := s.SyncBuildsPhotos(); err != nil {
			log.Println("SyncBuildsPhotos error - ", err.Error())
			return err
		}

		log.Println("Bucket " + s.BuildingsPhotosBucketName + "created")
	}

	found, err = s.Session.BucketExists(
		context.Background(),
		s.StaticFilesBucketName,
	)
	if err != nil {
		log.Fatalln(err)
	}

	if found {
		log.Println("Bucket found.")
	} else {
		log.Println("Bucket not found.")

		log.Println("Creating minio bucket start")

		opts := minio.MakeBucketOptions{
			ObjectLocking: false,
			Region:        "us-east-1",
		}

		err = s.Session.MakeBucket(
			context.Background(),
			s.StaticFilesBucketName,
			opts,
		)
		if err != nil {
			log.Fatalln("makebucket error - ", err.Error())
		}

		policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:ListBucket","s3:ListBucketMultipartUploads","s3:GetBucketLocation"],"Resource":["arn:aws:s3:::static"]},{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:AbortMultipartUpload","s3:DeleteObject","s3:GetObject","s3:ListMultipartUploadParts","s3:PutObject"],"Resource":["arn:aws:s3:::static/*"]}]}`

		err = s.Session.SetBucketPolicy(
			context.Background(),
			s.StaticFilesBucketName,
			policy,
		)
		if err != nil {
			log.Fatalln("SetBucketPolicy error - ", err.Error())
		}

		if err := s.SyncStaticFiles(); err != nil {
			log.Println("SyncStaticFiles error - ", err.Error())
			return err
		}

		log.Println("Bucket " + s.StaticFilesBucketName + "created")
	}

	return nil
}

func (s *MinioRepository) SyncBuildsPhotos() error {
	buildings, err := s.buildingsRepo.Buildings(context.Background())
	if err != nil {
		return err
	}

	buildsNames := map[string]string{}

	buildsNames[`Главный корпус`] = `0`
	buildsNames[`Учебно-лабораторный корпус`] = `1`
	buildsNames[`Корпус Э`] = `2`
	buildsNames[`Корпус СМ`] = `3`
	buildsNames[`Корпус Т`] = `4`

	for name, key := range buildsNames {
		if err := s.uploadPhoto(key, name, buildings); err != nil {
			fmt.Println(err.Error())
			continue
		}
	}

	return nil
}

func (s *MinioRepository) SyncStaticFiles() error {
	//object, err := os.Open(s.StaticFilesPath + "common.css")
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//defer func() {
	//	if err := object.Close(); err != nil {
	//		fmt.Println(err)
	//	}
	//}()
	//
	//objectStat, err := object.Stat()
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//info, err := s.Session.PutObject(
	//	context.Background(),
	//	s.StaticFilesBucketName,
	//	"common.css",
	//	object,
	//	objectStat.Size(),
	//	minio.PutObjectOptions{ContentType: "application/octet-stream"},
	//)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//log.Println(
	//	"Uploaded",
	//	"common.css",
	//	" of size: ",
	//	info.Size,
	//	"Successfully.",
	//)
	//
	//object, err := os.Open(s.StaticFilesPath + "common.css")
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//defer func() {
	//	if err := object.Close(); err != nil {
	//		fmt.Println(err)
	//	}
	//}()
	//
	//objectStat, err := object.Stat()
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//info, err := s.Session.PutObject(
	//	context.Background(),
	//	s.StaticFilesBucketName,
	//	"common.css",
	//	object,
	//	objectStat.Size(),
	//	minio.PutObjectOptions{ContentType: "application/octet-stream"},
	//)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//log.Println(
	//	"Uploaded",
	//	"common.css",
	//	" of size: ",
	//	info.Size,
	//	"Successfully.",
	//)

	return nil
}

func getBuildID(builds []model.BuildingModel, name string) (string, error) {
	for _, v := range builds {
		if strings.Contains(v.Name, name) {
			return v.Id, nil
		}
	}

	return "", errors.New("build not found")
}

func (s *MinioRepository) SaveBuildingPreview(
	ctx context.Context,
	id string,
	object io.Reader,
) error {
	//objectStat, err := object.Stat()
	//if err != nil {
	//	log.Fatalln(err)
	//}

	var ff multipart.File
	if _, err := s.Session.PutObject(
		ctx, s.BuildingsPhotosBucketName,
		id+".png",
		ff,
		1000000000,
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	); err != nil {
		return err
	}

	return nil
}

func (s *MinioRepository) uploadPhoto(
	key string,
	buildName string,
	buildings []model.BuildingModel,
) error {
	object, err := os.Open(s.PhotosLocalPath + key + ".png")
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := object.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	objectStat, err := object.Stat()
	if err != nil {
		log.Fatalln(err)
	}

	buildID, err := getBuildID(buildings, buildName)
	if err != nil {
		return err
	}

	info, err := s.Session.PutObject(
		context.Background(),
		s.BuildingsPhotosBucketName,
		buildID+".png",
		object,
		objectStat.Size(),
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(
		"Uploaded",
		buildID+".png",
		" of size: ",
		info.Size,
		"Successfully.",
	)

	if err := s.buildingsRepo.EditImgUrl(
		context.Background(),
		buildID,
		"/"+s.BuildingsPhotosBucketName+"/"+buildID+".png",
	); err != nil {
		log.Println(
			"Failed to update building image URL in postgres - ",
			err.Error(),
		)
	}

	return nil
}

func (s *MinioRepository) PrintBuilbingsBucketPolice() {
	policy, err := s.Session.GetBucketPolicy(
		context.Background(),
		s.BuildingsPhotosBucketName,
	)
	if err != nil {
		log.Fatalln(err)
	}

	log.Print("policy:")

	log.Print(policy)
}

func (s *MinioRepository) PrintStaticBucketPolice() {
	policy, err := s.Session.GetBucketPolicy(
		context.Background(),
		s.StaticFilesBucketName,
	)
	if err != nil {
		log.Fatalln(err)
	}

	log.Print("policy:")

	log.Print(policy)
}
