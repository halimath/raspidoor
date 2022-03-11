module github.com/halimath/raspidoor/cli

go 1.18

require (
	github.com/halimath/raspidoor/controller v0.0.0
	github.com/spf13/cobra v1.3.0
	google.golang.org/grpc v1.42.0
)

replace github.com/halimath/raspidoor/controller v0.0.0 => ../controller

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.0.0-20210813160813-60bc85c4be6d // indirect
	golang.org/x/sys v0.0.0-20211205182925-97ca703d548d // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20211208223120-3a66f561d7aa // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)
