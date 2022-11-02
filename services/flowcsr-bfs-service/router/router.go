package router

import (
	"bigrule/common/router"
	"bigrule/services/flowcsr-bfs-service/controller/parsers"
	"bigrule/services/flowcsr-bfs-service/controller/ping"
	"bigrule/services/flowcsr-bfs-service/controller/repos"
	"bigrule/services/flowcsr-bfs-service/controller/rules"
	"bigrule/services/flowcsr-bfs-service/controller/tags"
)

func RouterSetup() {
	router.RouterRegister(
		ping.PingRouter{},
		rules.RuleRouter{},
		tags.TagRouter{},
		parsers.ParserRouter{},
		repos.RepoRouter{},
	)
}
