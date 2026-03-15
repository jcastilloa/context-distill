package configui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TViewRunner struct {
	repository Repository
}

func NewTViewRunner(repository Repository) Runner {
	return TViewRunner{repository: repository}
}

func (r TViewRunner) Run(serviceName string) error {
	settings, err := r.repository.Load(serviceName)
	if err != nil {
		return err
	}

	app := tview.NewApplication()
	options := ProviderOptions()
	selectedIndex := providerIndex(options, settings.ProviderName)
	selectedProvider := options[selectedIndex]

	baseURL := settings.BaseURL
	if strings.TrimSpace(baseURL) == "" {
		baseURL = selectedProvider.DefaultBaseURL
	}

	header := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetText(
			"[#7dd3fc::b]Context Distill MCP Config[-:-:-]\n" +
				"[gray]Configuration is saved to ~/.config/" + serviceName + "/config.yaml",
		)

	statusView := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetText("[gray]Select provider, model, base URL, and API key when required.")

	helperView := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetText(providerHelperText(selectedProvider))

	providerDropDown := tview.NewDropDown().SetLabel("Provider   ")
	baseURLInput := tview.NewInputField().
		SetLabel("Base URL   ").
		SetText(baseURL).
		SetFieldWidth(64)
	apiKeyInput := tview.NewInputField().
		SetLabel("API Key    ").
		SetText(settings.APIKey).
		SetMaskCharacter('*').
		SetFieldWidth(64)
	model := settings.Model
	if strings.TrimSpace(model) == "" {
		model = selectedProvider.DefaultModel
	}
	modelInput := tview.NewInputField().
		SetLabel("Model      ").
		SetText(model).
		SetFieldWidth(64)

	optionLabels := make([]string, 0, len(options))
	for _, option := range options {
		optionLabels = append(optionLabels, option.Label)
	}
	providerDropDown.SetOptions(optionLabels, nil)
	providerDropDown.SetCurrentOption(selectedIndex)
	providerDropDown.SetSelectedFunc(func(_ string, index int) {
		if index < 0 || index >= len(options) {
			return
		}

		next := options[index]
		currentBaseURL := strings.TrimSpace(baseURLInput.GetText())
		if currentBaseURL == "" || currentBaseURL == selectedProvider.DefaultBaseURL {
			baseURLInput.SetText(next.DefaultBaseURL)
		}
		currentModel := strings.TrimSpace(modelInput.GetText())
		if currentModel == "" || currentModel == selectedProvider.DefaultModel {
			modelInput.SetText(next.DefaultModel)
		}

		helperView.SetText(providerHelperText(next))
		selectedProvider = next
	})

	form := tview.NewForm().
		AddFormItem(providerDropDown).
		AddFormItem(modelInput).
		AddFormItem(baseURLInput).
		AddFormItem(apiKeyInput).
		AddButton("Save", func() {
			index, _ := providerDropDown.GetCurrentOption()
			if index < 0 || index >= len(options) {
				statusView.SetText("[red]No provider selected.")
				return
			}

			saveSettings := DistillSettings{
				ProviderName: options[index].Name,
				Model:        modelInput.GetText(),
				BaseURL:      baseURLInput.GetText(),
				APIKey:       apiKeyInput.GetText(),
			}
			if err := ValidateSettings(saveSettings); err != nil {
				statusView.SetText("[red]" + err.Error())
				return
			}

			if err := r.repository.Save(serviceName, saveSettings); err != nil {
				statusView.SetText("[red]" + err.Error())
				return
			}

			statusView.SetText("[green]Configuration saved successfully.")
			app.Stop()
		}).
		AddButton("Cancel", func() {
			app.Stop()
		})

	form.SetButtonsAlign(tview.AlignRight)
	form.SetBorder(true).
		SetTitle("  MCP Setup  ").
		SetTitleAlign(tview.AlignLeft)
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetButtonBackgroundColor(tcell.ColorTeal)
	form.SetButtonTextColor(tcell.ColorBlack)

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(header, 2, 0, false).
		AddItem(form, 0, 1, true).
		AddItem(helperView, 2, 0, false).
		AddItem(statusView, 2, 0, false)

	layout.SetBorder(true)
	layout.SetBorderColor(tcell.ColorLightSkyBlue)
	layout.SetBackgroundColor(tcell.ColorBlack)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			app.Stop()
			return nil
		}
		return event
	})

	return app.SetRoot(layout, true).SetFocus(form).EnableMouse(true).Run()
}
