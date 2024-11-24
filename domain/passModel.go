package model

import (
	"fmt"
	"strconv"
	"time"
)

type PassModel struct {
	ID          int64
	Items       PassItems
	VisitorName string
	DateVisit   time.Time
}

type PassItem struct {
	Building *BuildingModel
	Comment  string
}

type PassItems []*PassItem

func (p *PassModel) GetHMTL(buildImagesHostname *string) *string {
	html := `
		<div id="cart-page-main-items">

            <ul id="cart-page-main-items-list">` +
		*p.Items.getHtml(buildImagesHostname) +
		`</ul>
            <div id="visitor">
                <h3>Посетитель:</h3>
                <p>` +
		p.VisitorName +
		`</p>
            </div>
            <div id="date">
                <h3>Дата:</h3>
                <p>` +
		p.DateVisit.String() +
		`</p>
            </div>
        </div>
		
		<div>
            <form action="delete_pass/` + strconv.Itoa(int(p.ID)) + `" method="post">
            <button type="submit">Удалить пропуск</button>
            </form>
        </div>
`

	return &html
}

func (p *PassItems) getHtml(buildImagesHostname *string) *string {
	var html string

	fmt.Println("getHtml passItems, элементов: ", len(*p))

	for _, passItem := range *p {
		html += `<li><div class="cart-item-img"><img src="` + *buildImagesHostname + passItem.Building.ImgUrl + `"></div><div class="cart-item-desc"><div>Услуга: ` + passItem.Building.Name + `</div><div>Описание: ` + passItem.Building.Description + `</div></div><div class="cart-item-quantity">Комментарий: ` + passItem.Comment + `</div></li>` + "\n"
	}

	fmt.Println("getHtml passItems, возвращаем: ", html)

	return &html
}
