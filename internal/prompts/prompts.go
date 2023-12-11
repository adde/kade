package prompts

import (
	"log"
	"os"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
)

func TextInput(label string, placeholder string, initialValue string, required bool) string {
	input := textinput.New(label)
	if placeholder != "" {
		input.Placeholder = placeholder
	}
	if initialValue != "" {
		input.InitialValue = initialValue
	}

	if !required {
		input.Validate = func(value string) error { return nil }
	}

	value, err := input.RunPrompt()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	return value
}

func SelectInput(label string, options []string) string {
	sp := selection.New(label, options)
	sp.Filter = nil

	value, err := sp.RunPrompt()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	return value
}

func ConfirmationInput(label string, initialValue confirmation.Value) bool {
	input := confirmation.New(label, initialValue)

	value, err := input.RunPrompt()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	return value
}

func PassWordInput(label string, placeholder string, required bool) string {
	input := textinput.New(label)
	input.Placeholder = placeholder
	input.Hidden = true

	if !required {
		input.Validate = func(value string) error { return nil }
	}

	value, err := input.RunPrompt()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	return value
}
