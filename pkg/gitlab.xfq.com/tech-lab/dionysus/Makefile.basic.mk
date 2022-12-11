PROJECT :=dionysus
DIR := /go/src/gitlab.xfq.com/tech-lab/dionysus

.PHONY: init
init:
	-@docker rm -f $(PROJECT)
	@docker pull reg.weipaitang.com/micro/golangci
	@docker run -td --name=$(PROJECT) \
        -v $(shell pwd):$(DIR) \
        -v ~/.ssh/id_rsa:/root/.ssh/id_rsa \
        -e GOSUMDB=sum.golang.google.cn \
        -e GONOSUMDB=gitlab.xfq.com \
        -e GOPROXY=http://nexus.wpt.la/repository/wpt-go-group/ \
		reg.weipaitang.com/micro/golangci /bin/bash


.PHONY: lint
lint:
	@docker start $(PROJECT)
	@docker exec -it $(PROJECT) bash -c "$(DIR)/hack/entry.sh && cd $(DIR) && go test -gcflags='-l' -v ./... -cover && lint golangci"
	@docker stop $(PROJECT)

.PHONY: goci
goci:
	@docker start $(PROJECT)
	@docker exec -it $(PROJECT) bash -c "cd $(DIR) && lint golangci"
	@docker stop $(PROJECT)
