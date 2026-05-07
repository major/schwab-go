module github.com/major/schwab-go/examples/auth

go 1.26

require (
	github.com/major/schwab-go v0.0.0
	github.com/major/schwab-go/schwab/auth v0.0.0
)

replace (
	github.com/major/schwab-go => ../../
	github.com/major/schwab-go/schwab/auth => ../../schwab/auth
)
