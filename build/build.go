package build

import (
	"bytes"
	"limn/render"
	"log"
	"os"
	"path/filepath"
)

func BuildSite(site *render.Renderer, siteDir string) {
    rndrfnc := func(path string,ctnt render.Content) error {
        buf := bytes.NewBuffer([]byte{})
        absPath := filepath.Join(siteDir, "build", path)
        dir, name := filepath.Split(absPath)
        if name == "index.html" { return nil }

        log.Printf("rendering %s", name)
        log.Printf("%s", absPath)
        ctnt.Render(buf, site.Template(), path)
        if dir != "" {
            log.Printf("making dir %s", dir)
            if err := os.MkdirAll(dir, 0777); err != nil {
                return err
            }
        }
        var err error
        switch filepath.Ext(name){
        case ".html":
            err = os.WriteFile(absPath, buf.Bytes(), 0777)
        case "":
            if err := os.MkdirAll(absPath, 0777); err != nil {
                return err
            }

            err = os.WriteFile(filepath.Join(absPath, "index.html"),
                buf.Bytes(), 0666)
        default:
            err = os.WriteFile(absPath, buf.Bytes(), 0777)
        }
        if err != nil {
            return err
        }
        
        return nil
    }
    os.RemoveAll(filepath.Join(siteDir, "build"))
    os.Mkdir(filepath.Join(siteDir, "build"), 0777)
    err := site.WalkSite(rndrfnc)
    if err != nil { log.Fatal(err) }
}

