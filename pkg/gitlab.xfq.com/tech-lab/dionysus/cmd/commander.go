package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Priority int

const (
	High = Priority(1)
	Mid  = High * 10
	Low  = Mid * 10
)

type Commander interface {
	GetCmd() *cobra.Command

	RegFlagSet(set *pflag.FlagSet)
	Flags() *pflag.FlagSet

	RegPreRunFunc(value string, priority Priority, f func() error) error
	RegPostRunFunc(value string, priority Priority, f func() error) error
}
