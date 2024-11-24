package handler_gin_v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"log"
	"net/http"
	"net/url"
	buildService "rip/service/build"
	passService "rip/service/pass"
	"strconv"
	"strings"
)

func MainPage(
	buildingsService *buildService.BuildService,
	passService *passService.PassService,
) func(
	c *gin.Context,
) {
	return func(c *gin.Context) {
		passID, err := passService.GetPassID("0")
		if err != nil {
			fmt.Println(err.Error())
		}

		if strings.Contains(c.Request.URL.String(), "/?buildName=") {
			//fmt.Println("вся строка:", c.Request.URL.String())
			//fmt.Println("параметр 11= ", c.Request.URL.String()[12:])
			decodedValue, err := url.QueryUnescape(c.Request.URL.String()[12:])
			if err != nil {
				fmt.Println(err.Error())
			}

			c.HTML(
				http.StatusOK, "mainPage.tmpl", gin.H{
					"services": template.HTML(
						buildingsService.GetBuilds(decodedValue),
					),
					"pass_id":   passID,
					"findValue": decodedValue,
				},
			)
			return
		}

		c.HTML(
			http.StatusOK, "mainPage.tmpl", gin.H{
				"services": template.HTML(
					buildingsService.GetBuilds(
						"",
					),
				),
				"pass_id": passID,
			},
		)
	}
}

func PassPage(
	passService *passService.PassService,
) func(
	c *gin.Context,
) {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))

		passHTML, err := passService.GetPassHTML(int64(id))
		if err != nil {
			c.Status(404)

			return
		}

		c.HTML(
			http.StatusOK, "passPage.tmpl", gin.H{
				"pass": template.HTML(
					*passHTML,
				),
			},
		)
	}
}

func BuildPage(
	buildingsService *buildService.BuildService,
) func(
	c *gin.Context,
) {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))

		building, err := buildingsService.GetBuild(
			int64(id),
		)

		if err != nil {
			log.Println(err.Error())
		}

		c.HTML(
			http.StatusOK, "servicePage.tmpl", gin.H{
				"name":        building.Name,
				"description": building.Description,
				"image":       template.HTML(`<img src="` + *buildingsService.GetBuildImagesHostname() + building.ImgUrl + `">`),
			},
		)
	}
}

func AddToPass(passService *passService.PassService) func(
	c *gin.Context,
) {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))

		passService.AddToPass(0, int64(id))

		fmt.Println("Add to Pass build = ", id)
		c.Redirect(http.StatusFound, "/")
	}
}

func DeletePass(passService *passService.PassService) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))

		passService.Delete(int64(id))

		fmt.Println("Delete pass, id = ", id)
		c.Redirect(http.StatusFound, "/")
	}
}
