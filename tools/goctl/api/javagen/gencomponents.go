package javagen

import (
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"text/template"

	"github.com/brucewang585/go-zero/tools/goctl/api/spec"
	apiutil "github.com/brucewang585/go-zero/tools/goctl/api/util"
	"github.com/brucewang585/go-zero/tools/goctl/util"
)

const (
	componentTemplate = `// Code generated by goctl. DO NOT EDIT.
package com.xhb.logic.http.packet.{{.packet}}.model;

import com.xhb.logic.http.DeProguardable;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

{{.componentType}}
`
)

func genComponents(dir, packetName string, api *spec.ApiSpec) error {
	types := apiutil.GetSharedTypes(api)
	if len(types) == 0 {
		return nil
	}
	for _, ty := range types {
		if err := createComponent(dir, packetName, ty, api.Types); err != nil {
			return err
		}
	}

	return nil
}

func createComponent(dir, packetName string, ty spec.Type, types []spec.Type) error {
	modelFile := util.Title(ty.Name) + ".java"
	filename := path.Join(dir, modelDir, modelFile)
	if err := util.RemoveOrQuit(filename); err != nil {
		return err
	}

	fp, created, err := apiutil.MaybeCreateFile(dir, modelDir, modelFile)
	if err != nil {
		return err
	}
	if !created {
		return nil
	}
	defer fp.Close()

	tys, err := buildType(ty, types)
	if err != nil {
		return err
	}

	t := template.Must(template.New("componentType").Parse(componentTemplate))
	return t.Execute(fp, map[string]string{
		"componentType": tys,
		"packet":        packetName,
	})
}

func buildType(ty spec.Type, types []spec.Type) (string, error) {
	var builder strings.Builder
	if err := writeType(&builder, ty, types); err != nil {
		return "", apiutil.WrapErr(err, "Type "+ty.Name+" generate error")
	}
	return builder.String(), nil
}

func writeType(writer io.Writer, tp spec.Type, types []spec.Type) error {
	fmt.Fprintf(writer, "public class %s implements DeProguardable {\n", util.Title(tp.Name))
	var members []spec.Member
	err := writeMembers(writer, types, tp.Members, &members, 1)
	if err != nil {
		return err
	}

	genGetSet(writer, members, 1)
	fmt.Fprintf(writer, "}")
	return nil
}

func writeMembers(writer io.Writer, types []spec.Type, members []spec.Member, allMembers *[]spec.Member, indent int) error {
	for _, member := range members {
		if !member.IsBodyMember() {
			continue
		}

		for _, item := range *allMembers {
			if item.Name == member.Name {
				continue
			}
		}

		if member.IsInline {
			hasInline := false
			for _, ty := range types {
				if strings.ToLower(ty.Name) == strings.ToLower(member.Name) {
					err := writeMembers(writer, types, ty.Members, allMembers, indent)
					if err != nil {
						return err
					}
					hasInline = true
					break
				}
			}
			if !hasInline {
				return errors.New("inline type " + member.Name + " not exist, please correct api file")
			}
		} else {
			if err := writeProperty(writer, member, indent); err != nil {
				return err
			}
			*allMembers = append(*allMembers, member)
		}
	}
	return nil
}
