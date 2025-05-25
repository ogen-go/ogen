package api

import (
	"encoding/json"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"
)

// TestMixedDiscriminatorMapping tests that the InlineQueryResult union properly handles
// both explicit discriminator mappings (like "sticker" -> InlineQueryResultCachedSticker)
// and implicit mappings (like "InlineQueryResultCachedAudio" -> InlineQueryResultCachedAudio)
// in the same union type.
func TestMixedDiscriminatorMapping(t *testing.T) {
	tests := []struct {
		name                string
		inputJSON           string
		expectedType        InlineQueryResultType
		expectedID          string
		isExplicitMapping   bool
		validateFunc        func(t *testing.T, result InlineQueryResult)
	}{
		{
			name:                "Explicit mapping: sticker",
			inputJSON:           `{"type":"sticker","id":"sticker-1","sticker_file_id":"file-1"}`,
			expectedType:        InlineQueryResultCachedStickerInlineQueryResult,
			expectedID:          "sticker-1",
			isExplicitMapping:   true,
			validateFunc: func(t *testing.T, result InlineQueryResult) {
				sticker, ok := result.GetInlineQueryResultCachedSticker()
				require.True(t, ok, "Should be cached sticker")
				require.Equal(t, "sticker-1", sticker.ID)
				require.Equal(t, "file-1", sticker.StickerFileID)
			},
		},
		{
			name:                "Implicit mapping: cached audio",
			inputJSON:           `{"type":"InlineQueryResultCachedAudio","id":"audio-1","audio_file_id":"file-2"}`,
			expectedType:        InlineQueryResultCachedAudioInlineQueryResult,
			expectedID:          "audio-1",
			isExplicitMapping:   false,
			validateFunc: func(t *testing.T, result InlineQueryResult) {
				cachedAudio, ok := result.GetInlineQueryResultCachedAudio()
				require.True(t, ok, "Should be cached audio")
				require.Equal(t, "audio-1", cachedAudio.ID)
				require.Equal(t, "file-2", cachedAudio.AudioFileID)
			},
		},
		{
			name:                "Explicit mapping: article",
			inputJSON:           `{"type":"article","id":"article-1","title":"Article","input_message_content":{"message_text":"Hello"}}`,
			expectedType:        InlineQueryResultArticleInlineQueryResult,
			expectedID:          "article-1",
			isExplicitMapping:   true,
			validateFunc: func(t *testing.T, result InlineQueryResult) {
				article, ok := result.GetInlineQueryResultArticle()
				require.True(t, ok, "Should be article")
				require.Equal(t, "article-1", article.ID)
				require.Equal(t, "Article", article.Title)
			},
		},
		{
			name:                "Implicit mapping: cached photo",
			inputJSON:           `{"type":"InlineQueryResultCachedPhoto","id":"photo-1","photo_file_id":"file-3"}`,
			expectedType:        InlineQueryResultCachedPhotoInlineQueryResult,
			expectedID:          "photo-1",
			isExplicitMapping:   false,
			validateFunc: func(t *testing.T, result InlineQueryResult) {
				cachedPhoto, ok := result.GetInlineQueryResultCachedPhoto()
				require.True(t, ok, "Should be cached photo")
				require.Equal(t, "photo-1", cachedPhoto.ID)
				require.Equal(t, "file-3", cachedPhoto.PhotoFileID)
			},
		},
		{
			name:                "Explicit mapping: audio",
			inputJSON:           `{"type":"audio","id":"audio-2","audio_url":"https://example.com/audio.mp3","title":"Audio"}`,
			expectedType:        InlineQueryResultAudioInlineQueryResult,
			expectedID:          "audio-2",
			isExplicitMapping:   true,
			validateFunc: func(t *testing.T, result InlineQueryResult) {
				audio, ok := result.GetInlineQueryResultAudio()
				require.True(t, ok, "Should be audio")
				require.Equal(t, "audio-2", audio.ID)
				require.Equal(t, "https://example.com/audio.mp3", audio.AudioURL.String())
				require.Equal(t, "Audio", audio.Title)
			},
		},
		{
			name:                "Implicit mapping: cached video",
			inputJSON:           `{"type":"InlineQueryResultCachedVideo","id":"video-1","video_file_id":"file-4","title":"Video"}`,
			expectedType:        InlineQueryResultCachedVideoInlineQueryResult,
			expectedID:          "video-1",
			isExplicitMapping:   false,
			validateFunc: func(t *testing.T, result InlineQueryResult) {
				cachedVideo, ok := result.GetInlineQueryResultCachedVideo()
				require.True(t, ok, "Should be cached video")
				require.Equal(t, "video-1", cachedVideo.ID)
				require.Equal(t, "file-4", cachedVideo.VideoFileID)
				require.Equal(t, "Video", cachedVideo.Title)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Decode individual JSON
			var result InlineQueryResult
			require.NoError(t, result.Decode(jx.DecodeBytes([]byte(tt.inputJSON))), 
				"Should decode %s mapping without error", 
				map[bool]string{true: "explicit", false: "implicit"}[tt.isExplicitMapping])

			// Verify discriminator type
			require.Equal(t, tt.expectedType, result.Type, 
				"Should have correct discriminator for %s mapping",
				map[bool]string{true: "explicit", false: "implicit"}[tt.isExplicitMapping])

			// Run custom validation
			tt.validateFunc(t, result)
		})
	}
}

// TestDiscriminatorMappingSettersPreserveExplicitMappings tests that when using
// setter methods, the discriminator values are set correctly for both explicit
// and implicit mappings.
func TestDiscriminatorMappingSettersPreserveExplicitMappings(t *testing.T) {
	tests := []struct {
		name                string
		setupFunc           func() InlineQueryResult
		expectedType        InlineQueryResultType
		expectedJSONType    string
		isExplicitMapping   bool
	}{
		{
			name: "Setter for explicit mapping preserves 'sticker' discriminator",
			setupFunc: func() InlineQueryResult {
				var result InlineQueryResult
				result.SetInlineQueryResultCachedSticker(InlineQueryResultCachedSticker{
					ID:           "sticker-id",
					StickerFileID: "file-id",
				})
				return result
			},
			expectedType:      InlineQueryResultCachedStickerInlineQueryResult,
			expectedJSONType:  "sticker",
			isExplicitMapping: true,
		},
		{
			name: "Setter for implicit mapping uses full type name",
			setupFunc: func() InlineQueryResult {
				var result InlineQueryResult
				result.SetInlineQueryResultCachedAudio(InlineQueryResultCachedAudio{
					ID:          "audio-id",
					AudioFileID: "file-id",
				})
				return result
			},
			expectedType:      InlineQueryResultCachedAudioInlineQueryResult,
			expectedJSONType:  "InlineQueryResultCachedAudio",
			isExplicitMapping: false,
		},
		{
			name: "Setter for explicit mapping preserves 'article' discriminator",
			setupFunc: func() InlineQueryResult {
				var result InlineQueryResult
				inputContent := InputTextMessageContent{
					MessageText: "Hello",
				}
				var msgContent InputMessageContent
				msgContent.SetInputTextMessageContent(inputContent)
				result.SetInlineQueryResultArticle(InlineQueryResultArticle{
					ID:                  "article-id",
					Title:               "Test Article",
					InputMessageContent: msgContent,
				})
				return result
			},
			expectedType:      InlineQueryResultArticleInlineQueryResult,
			expectedJSONType:  "article",
			isExplicitMapping: true,
		},
		{
			name: "Setter for implicit mapping uses full type name for cached photo",
			setupFunc: func() InlineQueryResult {
				var result InlineQueryResult
				result.SetInlineQueryResultCachedPhoto(InlineQueryResultCachedPhoto{
					ID:          "photo-id",
					PhotoFileID: "file-id",
				})
				return result
			},
			expectedType:      InlineQueryResultCachedPhotoInlineQueryResult,
			expectedJSONType:  "InlineQueryResultCachedPhoto",
			isExplicitMapping: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.setupFunc()

			// Verify discriminator is set correctly
			require.Equal(t, tt.expectedType, result.Type, "Discriminator should be set correctly")

			// Encode to JSON and verify the type field
			encoder := jx.Encoder{}
			result.Encode(&encoder)
			data := encoder.Bytes()

			// Parse just the type field to verify it matches expected
			var typeCheck struct {
				Type string `json:"type"`
			}
			require.NoError(t, json.Unmarshal(data, &typeCheck), "Should be able to decode type field")
			require.Equal(t, tt.expectedJSONType, typeCheck.Type, 
				"JSON type field should be '%s' for %s mapping", 
				tt.expectedJSONType,
				map[bool]string{true: "explicit", false: "implicit"}[tt.isExplicitMapping])

			// Verify round-trip preserves discriminator
			var decoded InlineQueryResult
			require.NoError(t, decoded.Decode(jx.DecodeBytes(data)), "Should decode without error")
			require.Equal(t, tt.expectedType, decoded.Type, "Discriminator should be preserved after round-trip")
		})
	}
}