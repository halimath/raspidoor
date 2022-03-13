module github.com/halimath/raspidoor/webapp

go 1.18

require (
	github.com/halimath/raspidoor/controller v0.0.0
	github.com/halimath/raspidoor/systemd v0.0.0
	google.golang.org/grpc v1.42.0
	gopkg.in/yaml.v2 v2.2.3
)

replace (
	github.com/halimath/raspidoor/controller v0.0.0 => ../controller
	github.com/halimath/raspidoor/systemd v0.0.0 => ../systemd
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	golang.org/x/net v0.0.0-20210813160813-60bc85c4be6d // indirect
	golang.org/x/sys v0.0.0-20211205182925-97ca703d548d // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20211208223120-3a66f561d7aa // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)
