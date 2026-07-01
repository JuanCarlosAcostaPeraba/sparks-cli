package cmd

import "context"

type optionsKey struct{}

func withRootOptions(ctx context.Context, opts *rootOptions) context.Context {
	return context.WithValue(ctx, optionsKey{}, opts)
}

func getRootOptions(cmd interface{ Context() context.Context }) *rootOptions {
	opts, _ := cmd.Context().Value(optionsKey{}).(*rootOptions)
	if opts == nil {
		return &rootOptions{}
	}
	return opts
}
