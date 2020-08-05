package msg

func init() {
	Processor.Register(&S2C_PayAccount{})

}

type S2C_PayAccount struct {
	Accounts []string
}