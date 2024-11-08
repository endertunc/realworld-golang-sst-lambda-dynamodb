package dto

import (
	"strings"
	"testing"
)

func TestCreateArticleRequestBodyDTO_Validate(t *testing.T) {
	tests := []ValidationTestCase[CreateArticleRequestBodyDTO]{
		{
			Name: "valid article request",
			Input: CreateArticleRequestBodyDTO{
				Article: CreateArticleRequestDTO{
					Title:       "Test Article",
					Description: "This is a test article",
					Body:        "Article body content",
					TagList:     []string{"test", "article"},
				},
			},
			WantErrors: false,
		},
		{
			Name: "missing title",
			Input: CreateArticleRequestBodyDTO{
				Article: CreateArticleRequestDTO{
					Description: "This is a test article",
					Body:        "Article body content",
					TagList:     []string{"test", "article"},
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"Article.Title": "Title is a required field",
			},
		},
		{
			Name: "blank title",
			Input: CreateArticleRequestBodyDTO{
				Article: CreateArticleRequestDTO{
					Title:       "    ",
					Description: "This is a test article",
					Body:        "Article body content",
					TagList:     []string{"test", "article"},
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"Article.Title": "Title cannot be blank",
			},
		},
		{
			Name: "title too long",
			Input: CreateArticleRequestBodyDTO{
				Article: CreateArticleRequestDTO{
					Title:       strings.Repeat("a", 256),
					Description: "This is a test article",
					Body:        "Article body content",
					TagList:     []string{"test", "article"},
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"Article.Title": "Title must be a maximum of 255 characters in length",
			},
		},
		{
			Name: "missing description",
			Input: CreateArticleRequestBodyDTO{
				Article: CreateArticleRequestDTO{
					Title:   "Test Article",
					Body:    "Article body content",
					TagList: []string{"test", "article"},
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"Article.Description": "Description is a required field",
			},
		},
		{
			Name: "blank description",
			Input: CreateArticleRequestBodyDTO{
				Article: CreateArticleRequestDTO{
					Title:       "Test Article",
					Description: "     ",
					Body:        "Article body content",
					TagList:     []string{"test", "article"},
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"Article.Description": "Description cannot be blank",
			},
		},
		{
			Name: "description too long",
			Input: CreateArticleRequestBodyDTO{
				Article: CreateArticleRequestDTO{
					Title:       "Test Article",
					Description: strings.Repeat("a", 1025),
					Body:        "Article body content",
					TagList:     []string{"test", "article"},
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"Article.Description": "Description must be a maximum of 1,024 characters in length",
			},
		},
		{
			Name: "missing body",
			Input: CreateArticleRequestBodyDTO{
				Article: CreateArticleRequestDTO{
					Title:       "Test Article",
					Description: "This is a test article",
					TagList:     []string{"test", "article"},
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"Article.Body": "Body is a required field",
			},
		},
		{
			Name: "blank body",
			Input: CreateArticleRequestBodyDTO{
				Article: CreateArticleRequestDTO{
					Title:       "Test Article",
					Description: "This is a test article",
					Body:        "     ",
					TagList:     []string{"test", "article"},
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"Article.Body": "Body cannot be blank",
			},
		},
		{
			Name: "empty tag list",
			Input: CreateArticleRequestBodyDTO{
				Article: CreateArticleRequestDTO{
					Title:       "Test Article",
					Description: "This is a test article",
					Body:        "Article body content",
					TagList:     []string{},
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"Article.TagList": "TagList must contain more than 0 items",
			},
		},
		{
			Name: "duplicate tags",
			Input: CreateArticleRequestBodyDTO{
				Article: CreateArticleRequestDTO{
					Title:       "Test Article",
					Description: "This is a test article",
					Body:        "Article body content",
					TagList:     []string{"test", "test"},
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"Article.TagList": "TagList must contain unique values",
			},
		},
		{
			Name: "blank tag",
			Input: CreateArticleRequestBodyDTO{
				Article: CreateArticleRequestDTO{
					Title:       "Test Article",
					Description: "This is a test article",
					Body:        "Article body content",
					TagList:     []string{"test", "   "},
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"Article.TagList[1]": "TagList[1] cannot be blank",
			},
		},
		{
			Name: "tag too long",
			Input: CreateArticleRequestBodyDTO{
				Article: CreateArticleRequestDTO{
					Title:       "Test Article",
					Description: "This is a test article",
					Body:        "Article body content",
					TagList:     []string{"test", strings.Repeat("a", 65)},
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"Article.TagList[1]": "TagList[1] must be a maximum of 64 characters in length",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			testValidation(t, tt)
		})
	}
}

func TestAddCommentRequestBodyDTO_Validate(t *testing.T) {
	tests := []ValidationTestCase[AddCommentRequestBodyDTO]{
		{
			Name: "valid comment request",
			Input: AddCommentRequestBodyDTO{
				Comment: AddCommentRequestDTO{
					Body: "This is a valid comment",
				},
			},
			WantErrors: false,
		},
		{
			Name: "missing body",
			Input: AddCommentRequestBodyDTO{
				Comment: AddCommentRequestDTO{},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"Comment.Body": "Body is a required field",
			},
		},
		{
			Name: "blank body",
			Input: AddCommentRequestBodyDTO{
				Comment: AddCommentRequestDTO{
					Body: "     ",
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"Comment.Body": "Body cannot be blank",
			},
		},
		{
			Name: "body too long",
			Input: AddCommentRequestBodyDTO{
				Comment: AddCommentRequestDTO{
					Body: strings.Repeat("a", 4097),
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"Comment.Body": "Body must be a maximum of 4,096 characters in length",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			testValidation(t, tt)
		})
	}
}
