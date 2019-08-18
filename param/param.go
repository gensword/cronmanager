package param

type AddJob struct {
	Name   string `json:"name" form:"name"` // job name
	Cmd    string `json:"command" form:"command"` // command to run
	Spec   string `json:"spec" form:"spec"` // job spec
	Status int    `json:"status" form:"status"` // 0 stop，1 run
}

type ModifyJob struct {
	Id int `json:"id"` // job id
	Name string `json:"name"`   // job name
	Cmd string `json:"command"` // command to run
	Spec string `json:"spec"`  // job spec
	Status int `json:"status"` // 0 stop, 1 run
}

type RemoveJob struct {
	Id int `json:"log_id"` // job id
}

type User struct {
	UserName string `json:"user_name"` // username
	Password string `json:"password"` // password
}
