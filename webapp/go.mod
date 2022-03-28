module github.com/halimath/raspidoor/webapp

go 1.18

require (
	github.com/halimath/appconf v0.0.0-20220327082622-fe7e06c8492d
	github.com/halimath/raspidoor/controller v0.0.0
	github.com/halimath/raspidoor/systemd v0.0.0
	google.golang.org/grpc v1.45.0
)

replace (
	github.com/halimath/raspidoor/controller v0.0.0 => ../controller
	github.com/halimath/raspidoor/systemd v0.0.0 => ../systemd
)

require (
	github.com/BurntSushi/toml v1.0.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	golang.org/x/net v0.0.0-20220325170049-de3da57026de // indirect
	golang.org/x/sys v0.0.0-20220325203850-36772127a21f // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220324131243-acbaeb5b85eb // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
