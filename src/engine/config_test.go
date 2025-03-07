package engine

import (
	"oh-my-posh/color"
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"oh-my-posh/segments"
	"testing"

	"github.com/gookit/config/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

func testClearDefaultConfig() {
	config.Default().ClearAll()
}

func TestParseMappedLocations(t *testing.T) {
	defer testClearDefaultConfig()
	cases := []struct {
		Case string
		JSON string
	}{
		{Case: "new format", JSON: `{ "properties": { "mapped_locations": {"folder1": "one","folder2": "two"} } }`},
		{Case: "old format", JSON: `{ "properties": { "mapped_locations": [["folder1", "one"], ["folder2", "two"]] } }`},
	}
	for _, tc := range cases {
		config.ClearAll()
		config.WithOptions(func(opt *config.Options) {
			opt.DecoderConfig = &mapstructure.DecoderConfig{
				TagName: "config",
			}
		})
		err := config.LoadStrings(config.JSON, tc.JSON)
		assert.NoError(t, err)
		var segment Segment
		err = config.BindStruct("", &segment)
		assert.NoError(t, err)
		mappedLocations := segment.Properties.GetKeyValueMap(segments.MappedLocations, make(map[string]string))
		assert.Equal(t, "two", mappedLocations["folder2"])
	}
}

func TestEscapeGlyphs(t *testing.T) {
	defer testClearDefaultConfig()
	cases := []struct {
		Input    string
		Expected string
	}{
		{Input: "a", Expected: "a"},
		{Input: "\ue0b4", Expected: "\\ue0b4"},
		{Input: "\ufd03", Expected: "\\ufd03"},
		{Input: "}", Expected: "}"},
		{Input: "🏚", Expected: "🏚"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.Expected, escapeGlyphs(tc.Input), tc.Input)
	}
}

func TestGetPalette(t *testing.T) {
	palette := color.Palette{
		"red":  "#ff0000",
		"blue": "#0000ff",
	}
	cases := []struct {
		Case            string
		Palettes        *color.Palettes
		Palette         color.Palette
		ExpectedPalette color.Palette
	}{
		{
			Case: "match",
			Palettes: &color.Palettes{
				Template: "{{ .Shell }}",
				List: map[string]color.Palette{
					"bash": palette,
					"zsh": {
						"red":  "#ff0001",
						"blue": "#0000fb",
					},
				},
			},
			ExpectedPalette: palette,
		},
		{
			Case: "no match, no fallback",
			Palettes: &color.Palettes{
				Template: "{{ .Shell }}",
				List: map[string]color.Palette{
					"fish": palette,
					"zsh": {
						"red":  "#ff0001",
						"blue": "#0000fb",
					},
				},
			},
			ExpectedPalette: nil,
		},
		{
			Case: "no match, default",
			Palettes: &color.Palettes{
				Template: "{{ .Shell }}",
				List: map[string]color.Palette{
					"zsh": {
						"red":  "#ff0001",
						"blue": "#0000fb",
					},
				},
			},
			Palette:         palette,
			ExpectedPalette: palette,
		},
		{
			Case:            "no palettes",
			ExpectedPalette: nil,
		},
	}
	for _, tc := range cases {
		env := &mock.MockedEnvironment{}
		env.On("TemplateCache").Return(&environment.TemplateCache{
			Env:   map[string]string{},
			Shell: "bash",
		})
		cfg := &Config{
			env:      env,
			Palette:  tc.Palette,
			Palettes: tc.Palettes,
		}
		got := cfg.getPalette()
		assert.Equal(t, tc.ExpectedPalette, got, tc.Case)
	}
}
