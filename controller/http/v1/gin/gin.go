package ginApi

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	model "rip/domain"
	cartRepository "rip/repository/carts/inmemory"
	serviceRepository "rip/repository/services/inmemory"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const ServiceImagesHostname = "http://localhost:9000"

func StartServer(
	servicesReporitory *serviceRepository.Repository,
	cartsRepository *cartRepository.Repository,
) {
	//log.Println("Server start up")

	r := gin.Default()

	//r.GET(
	//	"/ping", func(c *gin.Context) {
	//		c.JSON(
	//			http.StatusOK, gin.H{
	//				"message": "pong",
	//			},
	//		)
	//	},
	//)

	r.LoadHTMLGlob("templates/*")

	//services := getServices(servicesReporitory)
	//
	//fmt.Println(services)

	r.GET(
		"/", func(c *gin.Context) {
			if strings.Contains(c.Request.URL.String(), "/?serviceName=") {
				//fmt.Println("параметр = ", c.Request.URL.String()[14:])
				decodedValue, err := url.QueryUnescape(c.Request.URL.String()[14:])
				if err != nil {
					fmt.Println(err.Error())
				}
				c.HTML(
					http.StatusOK, "mainPage.tmpl", gin.H{
						"services": template.HTML(
							getServices(
								servicesReporitory,
								decodedValue,
							),
						),
						"cart_id":   getCartID(),
						"findValue": decodedValue,
					},
				)
				return
			}
			c.HTML(
				http.StatusOK, "mainPage.tmpl", gin.H{
					"services": template.HTML(
						getServices(
							servicesReporitory,
							"",
						),
					),
					"cart_id": getCartID(),
				},
			)
		},
	)

	r.GET(
		"/cart/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))

			cart := getCart(cartsRepository, int64(id))

			c.HTML(
				http.StatusOK, "cartPage.tmpl", gin.H{
					"items": template.HTML(
						getCartItems(
							servicesReporitory,
							cart.Items,
						),
					),
					"cost": cart.Cost,
				},
			)
		},
	)

	r.GET(
		"/services/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))

			service := getService(
				servicesReporitory,
				int64(id),
			)

			c.HTML(
				http.StatusOK, "servicePage.tmpl", gin.H{
					"name":        service.Name,
					"description": service.Description,
					"image":       template.HTML(`<img src="` + ServiceImagesHostname + service.ImgUrl + `">`),
					"button":      template.HTML(getButtonForDescriptionPage(int64(id))),
					"price":       strconv.Itoa(service.Price),
				},
			)
		},
	)

	r.Static("/image", "./static/images")
	r.Static("/style", "./static/styles")

	if err := r.Run(); err != nil {
		log.Fatalf(err.Error())
	}

	log.Println("Server down")
}

func getServices(r *serviceRepository.Repository, serviceName string) string {
	services := r.Services()

	var servicesHtml string

	if serviceName != "" {
		for _, service := range services {
			if strings.Contains(service.Name, serviceName) {
				serviceIdStr := strconv.Itoa(int(service.Id))
				serviceHtml := `<li><a href="/services/` + serviceIdStr + `" class="service-card">` + `<img src="` + ServiceImagesHostname + service.ImgUrl + `">` + `<h2>` + service.Name + `</h2><p>цена: ` + strconv.Itoa(service.Price) + `</p></a></li>` + "\n"
				servicesHtml += serviceHtml
			}
		}

		return servicesHtml
	}

	for _, service := range services {
		serviceIdStr := strconv.Itoa(int(service.Id))
		serviceHtml := `<li><a href="/services/` + serviceIdStr + `" class="service-card">` + `<img src="` + ServiceImagesHostname + service.ImgUrl + `">` + `<h2>` + service.Name + `</h2><p>цена: ` + strconv.Itoa(service.Price) + `</p></a></li>` + "\n"
		servicesHtml += serviceHtml
	}

	return servicesHtml
}

func getService(
	r *serviceRepository.Repository,
	serviceId int64,
) model.ServiceModel {
	service, _ := r.Service(serviceId)

	return service
}

func getCart(
	r *cartRepository.Repository,
	cartId int64,
) model.CartModel {
	cart, _ := r.Cart(cartId)

	return cart
}

func getButtonForDescriptionPage(serviceId int64) string {
	return `<button class="service-desc-btn">Добавить в корзину</button>`
}

func getCartItems(
	r *serviceRepository.Repository,
	items []model.ItemModel,
) string {
	var itemsHtml string

	for _, item := range items {
		service := getService(r, item.ServiceID)
		//serviceIdStr := strconv.Itoa(int(service.Id))
		itemHtml := `<li><div class="cart-item-img"><img src="` + ServiceImagesHostname + service.ImgUrl + `"></div><div class="cart-item-desc"><div>Услуга: ` + service.Name + `</div><div>Описание: ` + service.Description + `</div><div>Цена:` + strconv.Itoa(service.Price) + `</div></div><div class="cart-item-quantity">Количество: ` + strconv.Itoa(item.Quantity) + `</div></li>` + "\n"
		itemsHtml += itemHtml
	}

	return itemsHtml
}

func getCartID() string {
	return "0"
}
