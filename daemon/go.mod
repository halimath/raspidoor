module github.com/halimath/raspidoor/daemon

go 1.18

require (
	github.com/go-test/deep v1.0.8
	github.com/halimath/appconf v0.0.0-20220327082622-fe7e06c8492d
	github.com/halimath/raspidoor/controller v0.0.0
	github.com/halimath/raspidoor/systemd v0.0.0
	github.com/warthog618/gpiod v0.8.0
	google.golang.org/grpc v1.43.0
)

replace (
	github.com/halimath/raspidoor/controller v0.0.0 => ../controller
	github.com/halimath/raspidoor/systemd v0.0.0 => ../systemd
)

require (
	github.com/BurntSushi/toml v1.0.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/halimath/assertthat-go v0.0.0-20220327081729-20de7e695323 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	golang.org/x/net v0.0.0-20210813160813-60bc85c4be6d // indirect
	golang.org/x/sys v0.0.0-20220317061510-51cd9980dadf // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20211208223120-3a66f561d7aa // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
