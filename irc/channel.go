package irc

type Channel struct {
	Topic   string
	Members []*Member
	Mode    string
}

type Member struct {
	User  string
	Nick  string
	Oper  bool
	Voice bool
}
