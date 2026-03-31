package ai

type ImageCollectionPlan struct {
	ContentImageTasks []ImageSearchTask  `json:"contentImageTasks"`
	IllustrationTasks []IllustrationTask `json:"illustrationTasks"`
	DiagramTasks      []DiagramTask      `json:"diagramTasks"`
	LogoTasks         []LogoTask         `json:"logoTasks"`
}

type ImageSearchTask struct {
	Query string `json:"query"`
}

type IllustrationTask struct {
	Query string `json:"query"`
}

type DiagramTask struct {
	MermaidCode string `json:"mermaidCode"`
	Description string `json:"description"`
}

type LogoTask struct {
	Description string `json:"description"`
}
