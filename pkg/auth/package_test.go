package auth

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aler9/gortsplib/pkg/base"
	"github.com/aler9/gortsplib/pkg/headers"
)

func TestAuth(t *testing.T) {
	for _, c1 := range []struct {
		name    string
		methods []headers.AuthMethod
	}{
		{
			"basic",
			[]headers.AuthMethod{headers.AuthBasic},
		},
		{
			"digest",
			[]headers.AuthMethod{headers.AuthDigest},
		},
		{
			"both",
			[]headers.AuthMethod{headers.AuthBasic, headers.AuthDigest},
		},
	} {
		for _, conf := range []string{
			"nofail",
			"wronguser",
			"wrongpass",
			"wrongurl",
		} {
			if conf == "wrongurl" && c1.name == "basic" {
				continue
			}

			t.Run(c1.name+"_"+conf, func(t *testing.T) {
				va := NewValidator("testuser", "testpass", c1.methods)
				wwwAuthenticate := va.GenerateHeader()

				se, err := NewSender(wwwAuthenticate,
					func() string {
						if conf == "wronguser" {
							return "test1user"
						}
						return "testuser"
					}(),
					func() string {
						if conf == "wrongpass" {
							return "test1pass"
						}
						return "testpass"
					}())
				require.NoError(t, err)
				authorization := se.GenerateHeader(base.Announce,
					base.MustParseURL(func() string {
						if conf == "wrongurl" {
							return "rtsp://myhost/my1path"
						}
						return "rtsp://myhost/mypath"
					}()))

				err = va.ValidateHeader(authorization, base.Announce,
					base.MustParseURL("rtsp://myhost/mypath"))

				if conf != "nofail" {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			})
		}
	}
}

func TestAuthVLC(t *testing.T) {
	for _, ca := range []struct {
		clientURL string
		serverURL string
	}{
		{
			"rtsp://myhost/mypath/",
			"rtsp://myhost/mypath/trackID=0",
		},
		{
			"rtsp://myhost/mypath/test?testing/",
			"rtsp://myhost/mypath/test?testing/trackID=0",
		},
	} {
		va := NewValidator("testuser", "testpass",
			[]headers.AuthMethod{headers.AuthBasic, headers.AuthDigest})

		se, err := NewSender(va.GenerateHeader(), "testuser", "testpass")
		require.NoError(t, err)
		authorization := se.GenerateHeader(base.Announce,
			base.MustParseURL(ca.clientURL))

		err = va.ValidateHeader(authorization, base.Announce,
			base.MustParseURL(ca.serverURL))
		require.NoError(t, err)
	}
}

func TestAuthHashed(t *testing.T) {
	for _, conf := range []string{
		"nofail",
		"wronguser",
		"wrongpass",
	} {
		t.Run(conf, func(t *testing.T) {
			se := NewValidator("sha256:rl3rgi4NcZkpAEcacZnQ2VuOfJ0FxAqCRaKB/SwdZoQ=",
				"sha256:E9JJ8stBJ7QM+nV4ZoUCeHk/gU3tPFh/5YieiJp6n2w=",
				[]headers.AuthMethod{headers.AuthBasic, headers.AuthDigest})

			va, err := NewSender(se.GenerateHeader(),
				func() string {
					if conf == "wronguser" {
						return "test1user"
					}
					return "testuser"
				}(),
				func() string {
					if conf == "wrongpass" {
						return "test1pass"
					}
					return "testpass"
				}())
			require.NoError(t, err)
			authorization := va.GenerateHeader(base.Announce,
				base.MustParseURL("rtsp://myhost/mypath"))

			err = se.ValidateHeader(authorization, base.Announce,
				base.MustParseURL("rtsp://myhost/mypath"))

			if conf != "nofail" {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
