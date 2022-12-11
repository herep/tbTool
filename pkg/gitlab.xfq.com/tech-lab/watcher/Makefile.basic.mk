.PHONY: lint
DIR := /go/src/gitlab.xfq.com/tech-lab/watcher
lint:
	@docker pull reg.weipaitang.com/micro/golangci
	docker run -it --rm  \
        -v $(shell pwd):$(DIR) \
        -v ~/.ssh/id_rsa:/root/.ssh/id_rsa \
        reg.weipaitang.com/micro/golangci bash -c "$(DIR)/hack/entry.sh && cd $(DIR) && go test -v ./... -cover && lint golangci"
