module ohmnyom

go 1.16

require (
	cloud.google.com/go/firestore v1.6.1
	github.com/aiceru/protonyom v0.1.0
	github.com/stretchr/testify v1.7.0
	google.golang.org/api v0.65.0
	google.golang.org/grpc v1.43.0
)

replace github.com/aiceru/protonyom => ../protonyom
