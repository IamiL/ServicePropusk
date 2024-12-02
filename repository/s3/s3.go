package s3Repository

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func Connect(
	accessKeyID string,
	secretAccessKey string,
	bucketName string,
) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	//  КРИТИЧНО!  Важно заменить на 'localhost:9000' и указать endpoint.
	//  Без этого будет использоваться стандартный Amazon S3 endpoint, что не сработает!
	cfg.EndpointResolver = aws.EndpointResolverFunc(
		func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           "http://localhost:9000", //  ВАЖНО!
				SigningRegion: "us-east-1",             // or whatever your region is.  Не влияет на подключение, но  необходимо для aws-sdk
			}, nil
		},
	)

	client := s3.NewFromConfig(cfg)

	//const defaultRegion = "us-east-1"
	//staticResolver := aws.EndpointResolverFunc(
	//	func(service, region string) (aws.Endpoint, error) {
	//		return aws.Endpoint{
	//			PartitionID:       "aws",
	//			URL:               "http://localhost:9000", // or where ever you ran minio
	//			SigningRegion:     defaultRegion,
	//			HostnameImmutable: true,
	//		}, nil
	//	},
	//)
	//
	//cfg := aws.Config{
	//	Region: defaultRegion,
	//	Credentials: credentials.NewStaticCredentialsProvider(
	//		"minio124",
	//		"minio124",
	//		"",
	//	),
	//	EndpointResolver: staticResolver,
	//}
	//
	//client := s3.NewFromConfig(cfg)
	//
	//listObjectsOutput, err := client.ListObjectsV2(
	//	context.TODO(), &s3.ListObjectsV2Input{
	//		Bucket: &bucketName,
	//	},
	//)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//for _, object := range listObjectsOutput.Contents {
	//	obj, _ := json.MarshalIndent(object, "", "\t")
	//	fmt.Println(string(obj))
	//}
	//
	//listBucketsOutput, err := client.ListBuckets(
	//	context.TODO(),
	//	&s3.ListBucketsInput{},
	//)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//for _, object := range listBucketsOutput.Buckets {
	//	obj, _ := json.MarshalIndent(object, "", "\t")
	//	fmt.Println(string(obj))
	//}

	return client, nil
}

//func Connect2() *session.Session {
//	sdkConfig, err := config.LoadDefaultConfig(ctx)
//	s3Client := s3.NewFromConfig(sdkConfig)
//	AWS_S3_REGION := ""
//
//	sess, err := session.NewSession(
//		&aws.Config{
//			Region: aws.String(AWS_S3_REGION),
//		},
//	)
//
//	if err != nil {
//		panic(err)
//	}
//	return sess
//}

//type S3Repository struct {
//	Session                *session.Session
//	BuildsPhotosBucketName string
//	PhotosLocalPath        string
//}
//
//func New(
//	sess *session.Session,
//	buildsPhotosBucketName string,
//	PhotosLocalPath string,
//) *S3Repository {
//	return &S3Repository{sess, buildsPhotosBucketName, PhotosLocalPath}
//}
//
//func (s *S3Repository) SyncBuildsPhotos(
//	postgres *postgresBuilds.Storage,
//) error {
//	buildings, err := postgres.Buildings(context.Background())
//	if err != nil {
//		return err
//	}
//
//	buildsNames := map[string]string{}
//
//	buildsNames[`Главный корпус`] = `0`
//	buildsNames[`Учебно-лабораторный корпус`] = `1`
//	buildsNames[`Корпус Э`] = `2`
//	buildsNames[`Корпус СМ`] = `3`
//	buildsNames[`Корпус Т`] = `4`
//
//	uploader := s3manager.NewUploader(s.Session)
//
//	for name, key := range buildsNames {
//		if err := s.uploadPhoto(uploader, key, name, buildings); err != nil {
//			fmt.Println(err.Error())
//			continue
//		}
//	}
//
//	return nil
//}
//
//func getBuildID(builds []model.BuildingModel, name string) (string, error) {
//	for _, v := range builds {
//		if strings.Contains(v.Name, name) {
//			return v.Id, nil
//		}
//	}
//
//	return "", errors.New("build not found")
//}
//
//func (s *S3Repository) uploadPhoto(
//	uploader *s3manager.Uploader,
//	key string,
//	buildName string,
//	buildings []model.BuildingModel,
//) error {
//	file, err := os.Open(s.PhotosLocalPath + key + ".png")
//	defer func() {
//		if err := file.Close(); err != nil {
//			fmt.Println(err)
//		}
//	}()
//
//	if err != nil {
//		return err
//	}
//
//	reader := bufio.NewReader(file)
//
//	buildID, err := getBuildID(buildings, buildName)
//	if err != nil {
//		return err
//	}
//
//	if _, err := uploader.Upload(
//		&s3manager.UploadInput{
//			Bucket: aws.String(s.BuildsPhotosBucketName), // Bucket to be used
//			Key:    aws.String(buildID),                  // Name of the file to be saved
//			Body:   reader,                               // File
//		},
//	); err != nil {
//		return err
//	}
//
//	return nil
//}
