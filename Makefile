.PHONY: build serve test fmt clean

build:
	go run . build \
		-src testdata \
		-out build \
		-layout testdata/layout.html \
		-static testdata/static \
		-config config.yml \
		-posts testdata/posts

serve:
	go run . serve \
		-src testdata \
		-out build \
		-layout testdata/layout.html \
		-static testdata/static \
		-config config.yml \
		-posts testdata/posts \
		-addr :8080

test:
	go test ./...

fmt:
	gofmt -w .

clean:
	rm -rf build