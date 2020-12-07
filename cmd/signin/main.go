package signin

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/google/subcommands"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

type Command struct {
	username *string
	password *string

	clientId *string
	region   *string
}

func NewCommand() *Command {
	return &Command{}
}

func (c *Command) Name() string     { return "signin" }
func (c *Command) Synopsis() string { return "Sign in" }
func (c *Command) Usage() string {
	return `
`
}

func (c *Command) SetFlags(f *flag.FlagSet) {
	f.SetOutput(os.Stderr)

	c.username = f.String("u", "", "Username or email")
	c.password = f.String("p", "", "Password")
	c.clientId = f.String("client-id", "3sp4kn17md6qe043jqd3k1gf5u", "Cognito client ID")
	c.region = f.String("region", "ap-northeast-1", "region")
}

// Execute executes the command and returns an ExitStatus.
func (c *Command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	session := session.New(aws.NewConfig())
	provider := cognitoidentityprovider.New(session, aws.NewConfig().WithRegion(*c.region))

	res, err := provider.InitiateAuthWithContext(ctx, &cognitoidentityprovider.InitiateAuthInput{
		ClientId: aws.String(*c.clientId),
		AuthFlow: aws.String("USER_PASSWORD_AUTH"),
		AuthParameters: map[string]*string{
			"USERNAME": aws.String(*c.username),
			"PASSWORD": aws.String(*c.password),
		},
	})
	if err != nil {
		log.Fatalf("InitiateAuth: %v", err)
	}

	log.Println(res.GoString())

	return subcommands.ExitSuccess
}
