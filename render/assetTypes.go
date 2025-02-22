package render

type asset interface {
    acceptRenderer(rndr Renderer)
}

type file struct {
    
}

// func (rndr *Renderer) renderFile
