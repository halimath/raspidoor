module github.com/halimath/raspidoor/cli

go 1.18

require (
	github.com/halimath/raspidoor/controller v0.0.0
	github.com/spf13/cobra v1.4.0
	google.golang.org/grpc v1.45.0
)

replace github.com/halimath/raspidoor/controller v0.0.0 => ../controller

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.0.0-20220325170049-de3da57026de // indirect
	golang.org/x/sys v0.0.0-20220325203850-36772127a21f // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220324131243-acbaeb5b85eb // indirect
	google.golang.org/protobuf v1.28.0 // indirect
)
