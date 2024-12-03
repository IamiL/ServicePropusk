package handler_gin_v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"log"
	"net/http"
	"net/url"
	buildService "rip/internal/service/build"
	passService "rip/internal/service/pass"
	"strconv"
	"strings"
)

func MainPage(
	buildingsService *buildService.BuildingService,
	passService *passService.PassService,
) func(
	c *gin.Context,
) {
	return func(c *gin.Context) {
		passID, err := passService.GetPassID(c, "0")
		if err != nil {
			fmt.Println(err.Error())
		}

		passItemsCount, err := passService.GetPassItemsCount(c, "")

		if strings.Contains(c.Request.URL.String(), "/?buildName=") {
			decodedValue, err := url.QueryUnescape(c.Request.URL.String()[12:])
			if err != nil {
				fmt.Println(err.Error())
			}
			if decodedValue == "" {
				c.HTML(
					http.StatusOK, "mainPage.tmpl", gin.H{
						"services": template.HTML(
							buildingsService.GetAllBuildingsHTML(c),
						),
						"pass_id":          passID,
						"pass_items_count": strconv.Itoa(passItemsCount),
					},
				)
				return
			}

			c.HTML(
				http.StatusOK, "mainPage.tmpl", gin.H{
					"services": template.HTML(
						buildingsService.FindBuildings(c, decodedValue),
					),
					"pass_id":          passID,
					"findValue":        decodedValue,
					"pass_items_count": strconv.Itoa(passItemsCount),
				},
			)
			return
		}

		c.HTML(
			http.StatusOK, "mainPage.tmpl", gin.H{
				"services": template.HTML(
					buildingsService.GetAllBuildingsHTML(
						c,
					),
				),
				"pass_id":          passID,
				"pass_items_count": strconv.Itoa(passItemsCount),
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
		id := c.Param("id")

		passHTML, err := passService.GetPassHTML(c, id)
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

func BuildingPage(
	buildingsService *buildService.BuildingService,
) func(
	c *gin.Context,
) {
	return func(c *gin.Context) {
		id := c.Param("id")

		building, err := buildingsService.GetBuilding(
			c, id,
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
		id := c.Param("id")

		passService.AddToPass(c, "", id)

		fmt.Println("Add to Pass build = ", id)
		c.Redirect(http.StatusFound, "/")
	}
}

func DeletePass(passService *passService.PassService) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")

		passService.Delete(c, id)

		fmt.Println("Delete pass, id = ", id)
		c.Redirect(http.StatusFound, "/")
	}
}
