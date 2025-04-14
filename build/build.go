package build

import (
	"bytes"
	"limn/render"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func BuildSite(rndr *render.Renderer, siteDir string) {
    rndrfnc := func(path string,ctnt render.Content) error {
        url := render.URL(path)
        absPath := filepath.Join(siteDir, "build", url.Path())
        dir, name := filepath.Split(absPath)
        buf := bytes.NewBuffer([]byte{})

        if name == "index.html" { return nil }

        log.Printf("rendering %s", url.NoteName())
        log.Printf("%s", absPath)
        if filepath.Ext(url.Path()) == ".html" ||
            filepath.Ext(url.Path()) == "" {
            rndr.Site().Render(buf, rndr.Template(), url.Path())
        } else {
            // return nil
            err := rndr.URL(buf, url.Path())
            if err != nil {
                return err
            }
        }

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
		err = runScript(absPath, rndr.Config().Scripts)
		if err != nil {
			log.Fatalf("err while running script %s: %s",
				rndr.Config().Scripts[filepath.Ext(path)],
				err)
		}
        return nil
    }
    os.RemoveAll(filepath.Join(siteDir, "build"))
    os.Mkdir(filepath.Join(siteDir, "build"), 0777)
    err := rndr.WalkSite(rndrfnc)
    if err != nil { log.Fatal(err) }
}

func runScript(path string, scripts map[string]string) error {
	script, ok := scripts[filepath.Ext(path)]
	if !ok { return nil }
	cmd := exec.Command(script, path)
	err := cmd.Run()
	return err
}
