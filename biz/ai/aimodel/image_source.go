package ai

type ImageCategoryEnum string

const (
	ImageCategoryContent      ImageCategoryEnum = "CONTENT"
	ImageCategoryLogo         ImageCategoryEnum = "LOGO"
	ImageCategoryIllustration ImageCategoryEnum = "ILLUSTRATION"
	ImageCategoryArchitecture ImageCategoryEnum = "ARCHITECTURE"
)

func (e ImageCategoryEnum) Text() string {
	switch e {
	case ImageCategoryContent:
		return "内容图片"
	case ImageCategoryLogo:
		return "LOGO图片"
	case ImageCategoryIllustration:
		return "插画图片"
	case ImageCategoryArchitecture:
		return "架构图片"
	default:
		return ""
	}
}

func (e ImageCategoryEnum) Value() string {
	return string(e)
}

func GetImageCategoryByValue(value string) ImageCategoryEnum {
	if value == "" {
		return ""
	}
	return ImageCategoryEnum(value)
}

type ImageSource struct {
	Category    ImageCategoryEnum
	Description string
	Url         string
}

func NewImageSource(category ImageCategoryEnum, description string, url string) *ImageSource {
	return &ImageSource{
		Category:    category,
		Description: description,
		Url:         url,
	}
}
