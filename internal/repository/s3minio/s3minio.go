package s3minio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	model "service-propusk-backend/internal/domain"
	postgresBuilds "service-propusk-backend/internal/repository/postgres/buildings"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Host                      string `yaml:"host"`
	Port                      string `yaml:"port"`
	AccessKey                 string `yaml:"accessKey"`
	SecretKey                 string `yaml:"secretKey"`
	BuildingsPhotosBucketName string `yaml:"buildings_photos_bucket_name"`
	StaticFilesBucketName     string `yaml:"static_files_bucket_name"`
	QRCodesBucketName         string `yaml:"qr_codes_bucket_name"`
	BuildsPhotosLocalPath     string `yaml:"builds_photos_local_path"`
	StaticFilesLocalPath      string `yaml:"static_files_local_path"`
}

func (c *Config) Endpoint() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func NewConn(
	config *Config,
) (*minio.Client, error) {
	useSSL := false

	// Initialize minio client object.
	minioClient, err := minio.New(
		config.Endpoint(), &minio.Options{
			Creds: credentials.NewStaticV4(
				config.AccessKey,
				config.SecretKey,
				"",
			),
			Secure: useSSL,
		},
	)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = minioClient.ListBuckets(context.Background())
	if err != nil {
		return nil, err
	}

	return minioClient, nil
}

type MinioRepository struct {
	Session                   *minio.Client
	BuildingsPhotosBucketName string
	StaticFilesBucketName     string
	QRCodesBucketName         string
	PhotosLocalPath           string
	StaticFilesPath           string
	buildingsRepo             *postgresBuilds.Storage
}

func New(
	sess *minio.Client,
	buildsPhotosBucketName string,
	staticFilesBucketName string,
	qrCodesBucketName string,
	photosLocalPath string,
	staticFilesPath string,
	buildingsRepo *postgresBuilds.Storage,
) *MinioRepository {
	return &MinioRepository{
		sess,
		buildsPhotosBucketName,
		staticFilesBucketName,
		qrCodesBucketName,
		photosLocalPath,
		staticFilesPath,
		buildingsRepo,
	}
}

func (s *MinioRepository) ConfigureMinioStorage() error {
	// Configure buildings photos bucket
	found, err := s.Session.BucketExists(
		context.Background(),
		s.BuildingsPhotosBucketName,
	)
	if err != nil {
		log.Fatalln(err)
	}

	if found {
		log.Println("Buildings photos bucket found.")
	} else {
		log.Println("Buildings photos bucket not found.")

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

		log.Println("Bucket " + s.BuildingsPhotosBucketName + " created")
	}

	// Configure static files bucket
	found, err = s.Session.BucketExists(
		context.Background(),
		s.StaticFilesBucketName,
	)
	if err != nil {
		log.Fatalln(err)
	}

	if found {
		log.Println("Static files bucket found.")
	} else {
		log.Println("Static files bucket not found.")

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

		log.Println("Bucket " + s.StaticFilesBucketName + " created")
	}

	// Configure QR codes bucket
	found, err = s.Session.BucketExists(
		context.Background(),
		s.QRCodesBucketName,
	)
	if err != nil {
		log.Fatalln(err)
	}

	if found {
		log.Println("QR codes bucket found.")
	} else {
		log.Println("QR codes bucket not found.")

		log.Println("Creating QR codes bucket")

		opts := minio.MakeBucketOptions{
			ObjectLocking: false,
			Region:        "us-east-1",
		}

		err = s.Session.MakeBucket(
			context.Background(),
			s.QRCodesBucketName,
			opts,
		)
		if err != nil {
			log.Fatalln("makebucket error - ", err.Error())
		}

		policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:ListBucket","s3:ListBucketMultipartUploads","s3:GetBucketLocation"],"Resource":["arn:aws:s3:::qrcodes"]},{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:AbortMultipartUpload","s3:DeleteObject","s3:GetObject","s3:ListMultipartUploadParts","s3:PutObject"],"Resource":["arn:aws:s3:::qrcodes/*"]}]}`

		err = s.Session.SetBucketPolicy(
			context.Background(),
			s.QRCodesBucketName,
			policy,
		)
		if err != nil {
			log.Fatalln("SetBucketPolicy error - ", err.Error())
		}

		log.Println("Bucket " + s.QRCodesBucketName + " created")
	}

	return nil
}

func (s *MinioRepository) SyncBuildsPhotos() error {
	buildings, err := s.buildingsRepo.AllBuildings(context.Background())
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
	//		return err
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
	object []byte,
) error {
	log.Printf("saving building preview: id=%s, size=%d bytes", id, len(object))

	reader := bytes.NewReader(object)
	info, err := s.Session.PutObject(
		ctx,
		s.BuildingsPhotosBucketName,
		id+".png",
		reader,
		int64(len(object)),
		minio.PutObjectOptions{
			ContentType: "image/png",
		},
	)
	if err != nil {
		log.Printf("failed to save building preview: %v", err)
		return fmt.Errorf("failed to save building preview: %w", err)
	}

	log.Printf(
		"successfully saved building preview: id=%s, size=%d bytes",
		id,
		info.Size,
	)
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

	if err := s.buildingsRepo.EditBuildingImgUrl(
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

func (s *MinioRepository) DeleteBuildingPreview(
	ctx context.Context,
	id string,
) error {
	opts := minio.RemoveObjectOptions{
		GovernanceBypass: true,
	}
	err := s.Session.RemoveObject(
		context.Background(),
		s.BuildingsPhotosBucketName,
		id+`.png`,
		opts,
	)
	if err != nil {
		fmt.Println()
		return err
	}

	return nil
}

// SaveQRCode saves a QR code image to the QR codes bucket
func (s *MinioRepository) SaveQRCode(
	ctx context.Context,
	id string,
	qrCode []byte,
) error {
	log.Printf("saving QR code: id=%s, size=%d bytes", id, len(qrCode))

	reader := bytes.NewReader(qrCode)
	info, err := s.Session.PutObject(
		ctx,
		s.QRCodesBucketName,
		id+".png",
		reader,
		int64(len(qrCode)),
		minio.PutObjectOptions{
			ContentType: "image/png",
		},
	)
	if err != nil {
		log.Printf("failed to save QR code: %v", err)
		return fmt.Errorf("failed to save QR code: %w", err)
	}

	log.Printf(
		"successfully saved QR code: id=%s, size=%d bytes",
		id,
		info.Size,
	)
	return nil
}
