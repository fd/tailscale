// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

// get-authkey allocates an authkey using an OAuth API client
// https://tailscale.com/kb/1215/oauth-clients/ and prints it
// to stdout for scripts to capture and use.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/oauth2/clientcredentials"
	"tailscale.com/client/tailscale"
)

func main() {
	// Required to use our client API. We're fine with the instability since the
	// client lives in the same repo as this code.
	tailscale.I_Acknowledge_This_API_Is_Unstable = true

	reusable := flag.Bool("reusable", false, "allocate a reusable authkey")
	ephemeral := flag.Bool("ephemeral", false, "allocate an ephemeral authkey")
	preauth := flag.Bool("preauth", true, "set the authkey as pre-authorized")
	tags := flag.String("tags", "", "comma-separated list of tags to apply to the authkey")
	flag.Parse()

	clientId := os.Getenv("TS_API_CLIENT_ID")
	clientSecret := os.Getenv("TS_API_CLIENT_SECRET")
	if clientId == "" || clientSecret == "" {
		log.Fatal("TS_API_CLIENT_ID and TS_API_CLIENT_SECRET must be set")
	}

	baseUrl := os.Getenv("TS_BASE_URL")
	if baseUrl == "" {
		baseUrl = "https://api.tailscale.com"
	}

	credentials := clientcredentials.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		TokenURL:     baseUrl + "/api/v2/oauth/token",
		Scopes:       []string{"device"},
	}

	ctx := context.Background()
	tsClient := tailscale.NewClient("-", nil)
	tsClient.HTTPClient = credentials.Client(ctx)
	tsClient.BaseURL = baseUrl

	caps := tailscale.KeyCapabilities{
		Devices: tailscale.KeyDeviceCapabilities{
			Create: tailscale.KeyDeviceCreateCapabilities{
				Reusable:      *reusable,
				Ephemeral:     *ephemeral,
				Preauthorized: *preauth,
				Tags:          strings.Split(*tags, ","),
			},
		},
	}

	authkey, _, err := tsClient.CreateKey(ctx, caps)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println(authkey)
}