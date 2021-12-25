package main

import (
	"fmt"

	"github.com/opensourceways/community-robot-lib/config"
)

type configuration struct {
	ConfigItems []botConfig `json:"config_items,omitempty"`
}

func (c *configuration) configFor(org, repo string) *botConfig {
	if c == nil {
		return nil
	}

	items := c.ConfigItems
	v := make([]config.IRepoFilter, len(items))
	for i := range items {
		v[i] = &items[i]
	}

	if i := config.Find(org, repo, v); i >= 0 {
		return &items[i]
	}

	return nil
}

func (c *configuration) Validate() error {
	if c == nil {
		return nil
	}

	items := c.ConfigItems
	for i := range items {
		if err := items[i].validate(); err != nil {
			return err
		}
	}

	return nil
}

func (c *configuration) SetDefault() {
	if c == nil {
		return
	}

	Items := c.ConfigItems
	for i := range Items {
		Items[i].setDefault()
	}
}

type botConfig struct {
	config.RepoFilter

	CI ciConfig `json:"ci"`

	Review reviewConfig `json:"review"`

	Owner ownerConfig `json:"owner"`

	// NeedWelcome specifies whether to add welcome comment.
	NeedWelcome bool `json:"need_welcome,omitempty"`

	// Doc describes useful information about review process of PR.
	Doc string `json:"doc,omitempty"`
}

func (c *botConfig) setDefault() {
	if c != nil {
		c.CI.setDefault()
		c.Review.setDefault()
		c.Owner.setDefault()
	}
}

func (c *botConfig) validate() error {
	if c == nil {
		return nil
	}

	if err := c.CI.validate(); err != nil {
		return err
	}

	if err := c.Review.validate(); err != nil {
		return err
	}

	if err := c.Owner.validate(); err != nil {
		return err
	}

	if c.NeedWelcome && c.Doc == "" {
		return fmt.Errorf("missing doc")
	}

	return c.RepoFilter.Validate()
}
