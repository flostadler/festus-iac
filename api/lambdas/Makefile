build-api::
	GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o out/api/bootstrap cmd/api/main.go

.PHONY: out/api/bootstrap
out/api/bootstrap:
	GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o out/api/bootstrap cmd/api/main.go

.PHONY: out/account-update/bootstrap
out/account-update/bootstrap:
	GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o out/account-update/bootstrap cmd/account-update/main.go

out/api/function.zip: out/api/bootstrap
	cd out/api; zip function.zip bootstrap

out/account-update/function.zip: out/account-update/bootstrap
	cd out/account-update; zip function.zip bootstrap

.PHONY: clean
clean:
	rm -rf out

.PHONY: package
package: out/api/function.zip out/account-update/function.zip
