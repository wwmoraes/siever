package managesieve

import (
	"reflect"
	"testing"
)

func TestCapabilities_Parse(t *testing.T) {
	type args struct {
		content []string
	}
	tests := []struct {
		name    string
		args    args
		want    *Capabilities
		wantErr bool
	}{
		{
			name: "all values",
			args: args{
				content: []string{
					`"IMPLEMENTATION" "Dovecot Pigeonhole"`,
					`"SIEVE" "fileinto reject envelope encoded-character vacation subaddress comparator-i;ascii-numeric relational regex imap4flags copy include variables body enotify environment mailbox date index ihave duplicate mime foreverypart extracttext vacation-seconds editheader imapflags notify imapsieve vnd.dovecot.imapsieve vnd.dovecot.pgp-encrypt"`,
					`"MAXREDIRECTS" "40"`,
					`"NOTIFY" "mailto"`,
					`"SASL" "PLAIN LOGIN"`,
					`"VERSION" "1.0"`,
				},
			},
			want: &Capabilities{
				Implementation: "Dovecot Pigeonhole",
				Sieve:          []string{"fileinto", "reject", "envelope", "encoded-character", "vacation", "subaddress", "comparator-i;ascii-numeric", "relational", "regex", "imap4flags", "copy", "include", "variables", "body", "enotify", "environment", "mailbox", "date", "index", "ihave", "duplicate", "mime", "foreverypart", "extracttext", "vacation-seconds", "editheader", "imapflags", "notify", "imapsieve", "vnd.dovecot.imapsieve", "vnd.dovecot.pgp-encrypt"},
				MaxRedirects:   40,
				Notify:         []string{"mailto"},
				Version:        "1.0",
				SASL:           []string{"PLAIN", "LOGIN"},
				StartTLS:       false,
				Language:       "",
				Owner:          "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capabilities, err := ParseCapabilities(tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Capabilities.Parse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.want, capabilities) {
				t.Errorf("got %#+v, want %#+v", capabilities, tt.want)
			}
		})
	}
}
