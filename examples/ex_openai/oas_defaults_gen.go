// Code generated by ogen, DO NOT EDIT.

package api

// setDefaults set default value of fields.
func (s *CreateAnswerRequest) setDefaults() {
	{
		val := string("ada")
		s.SearchModel.SetTo(val)
	}
	{
		val := int(200)
		s.MaxRerank.SetTo(val)
	}
	{
		val := float64(0)
		s.Temperature.SetTo(val)
	}
	{
		s.Logprobs.Null = true
	}
	{
		val := int(16)
		s.MaxTokens.SetTo(val)
	}
	{
		val := int(1)
		s.N.SetTo(val)
	}
	{
		val := bool(false)
		s.ReturnMetadata.SetTo(val)
	}
	{
		val := bool(false)
		s.ReturnPrompt.SetTo(val)
	}
}

// setDefaults set default value of fields.
func (s *CreateChatCompletionRequest) setDefaults() {
	{
		val := float64(1)
		s.Temperature.SetTo(val)
	}
	{
		val := float64(1)
		s.TopP.SetTo(val)
	}
	{
		val := int(1)
		s.N.SetTo(val)
	}
	{
		val := bool(false)
		s.Stream.SetTo(val)
	}
	{
		val := float64(0)
		s.PresencePenalty.SetTo(val)
	}
	{
		val := float64(0)
		s.FrequencyPenalty.SetTo(val)
	}
}

// setDefaults set default value of fields.
func (s *CreateClassificationRequest) setDefaults() {
	{
		val := string("ada")
		s.SearchModel.SetTo(val)
	}
	{
		val := float64(0)
		s.Temperature.SetTo(val)
	}
	{
		s.Logprobs.Null = true
	}
	{
		val := int(200)
		s.MaxExamples.SetTo(val)
	}
	{
		val := bool(false)
		s.ReturnPrompt.SetTo(val)
	}
	{
		val := bool(false)
		s.ReturnMetadata.SetTo(val)
	}
}

// setDefaults set default value of fields.
func (s *CreateCompletionRequest) setDefaults() {
	{
		s.Suffix.Null = true
	}
	{
		val := int(16)
		s.MaxTokens.SetTo(val)
	}
	{
		val := float64(1)
		s.Temperature.SetTo(val)
	}
	{
		val := float64(1)
		s.TopP.SetTo(val)
	}
	{
		val := int(1)
		s.N.SetTo(val)
	}
	{
		val := bool(false)
		s.Stream.SetTo(val)
	}
	{
		s.Logprobs.Null = true
	}
	{
		val := bool(false)
		s.Echo.SetTo(val)
	}
	{
		val := float64(0)
		s.PresencePenalty.SetTo(val)
	}
	{
		val := float64(0)
		s.FrequencyPenalty.SetTo(val)
	}
	{
		val := int(1)
		s.BestOf.SetTo(val)
	}
}

// setDefaults set default value of fields.
func (s *CreateEditRequest) setDefaults() {
	{
		val := string("")
		s.Input.SetTo(val)
	}
	{
		val := int(1)
		s.N.SetTo(val)
	}
	{
		val := float64(1)
		s.Temperature.SetTo(val)
	}
	{
		val := float64(1)
		s.TopP.SetTo(val)
	}
}

// setDefaults set default value of fields.
func (s *CreateFineTuneRequest) setDefaults() {
	{
		val := string("curie")
		s.Model.SetTo(val)
	}
	{
		val := int(4)
		s.NEpochs.SetTo(val)
	}
	{
		s.BatchSize.Null = true
	}
	{
		s.LearningRateMultiplier.Null = true
	}
	{
		val := float64(0.01)
		s.PromptLossWeight.SetTo(val)
	}
	{
		val := bool(false)
		s.ComputeClassificationMetrics.SetTo(val)
	}
	{
		s.ClassificationNClasses.Null = true
	}
	{
		s.ClassificationPositiveClass.Null = true
	}
	{
		s.Suffix.Null = true
	}
}

// setDefaults set default value of fields.
func (s *CreateImageEditRequestMultipart) setDefaults() {
	{
		val := int(1)
		s.N.SetTo(val)
	}
	{
		val := CreateImageEditRequestMultipartSize("1024x1024")
		s.Size.SetTo(val)
	}
	{
		val := CreateImageEditRequestMultipartResponseFormat("url")
		s.ResponseFormat.SetTo(val)
	}
}

// setDefaults set default value of fields.
func (s *CreateImageRequest) setDefaults() {
	{
		val := int(1)
		s.N.SetTo(val)
	}
	{
		val := CreateImageRequestSize("1024x1024")
		s.Size.SetTo(val)
	}
	{
		val := CreateImageRequestResponseFormat("url")
		s.ResponseFormat.SetTo(val)
	}
}

// setDefaults set default value of fields.
func (s *CreateImageVariationRequestMultipart) setDefaults() {
	{
		val := int(1)
		s.N.SetTo(val)
	}
	{
		val := CreateImageVariationRequestMultipartSize("1024x1024")
		s.Size.SetTo(val)
	}
	{
		val := CreateImageVariationRequestMultipartResponseFormat("url")
		s.ResponseFormat.SetTo(val)
	}
}

// setDefaults set default value of fields.
func (s *CreateModerationRequest) setDefaults() {
	{
		val := string("text-moderation-latest")
		s.Model.SetTo(val)
	}
}

// setDefaults set default value of fields.
func (s *CreateSearchRequest) setDefaults() {
	{
		val := int(200)
		s.MaxRerank.SetTo(val)
	}
	{
		val := bool(false)
		s.ReturnMetadata.SetTo(val)
	}
}

// setDefaults set default value of fields.
func (s *CreateTranscriptionRequestMultipart) setDefaults() {
	{
		val := string("json")
		s.ResponseFormat.SetTo(val)
	}
	{
		val := float64(0)
		s.Temperature.SetTo(val)
	}
}

// setDefaults set default value of fields.
func (s *CreateTranslationRequestMultipart) setDefaults() {
	{
		val := string("json")
		s.ResponseFormat.SetTo(val)
	}
	{
		val := float64(0)
		s.Temperature.SetTo(val)
	}
}
