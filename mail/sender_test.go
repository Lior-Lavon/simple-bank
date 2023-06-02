package mail

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/liorlavon/simplebank/util"
)

func TestEmailWithGmail(t *testing.T) {

	// skipping the test during build process test phase
	if testing.Short() { // skipp this test if Short flag is set
		t.Skip() // set the short flag in the Make file
	}

	// load the config
	config, err := util.LoadConfig("..")
	require.NoError(t, err)

	// create email sender
	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	subject := "test email"
	content := `
	<h1>Test email</h1>
	<p>This is a test message from <a href="http://techschool.guru">Tech School</p>
	`
	to := []string{"liorlavon554@gmail.com"}
	attachFiles := []string{"../README.md"}

	err = sender.SendEmail(subject, content, to, nil, nil, attachFiles)
	require.NoError(t, err)
}
