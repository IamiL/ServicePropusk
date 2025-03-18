package model

import (
	"time"
)

type PassModel struct {
	ID          string
	Status      int
	CreatorID   string
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
            <form action="delete_pass/` + p.ID + `" method="post">
            <button type="submit">Удалить пропуск</button>
            </form>
        </div>
`

	return &html
}

func (p *PassItems) getHtml(buildImagesHostname *string) *string {
	var html string

	for _, passItem := range *p {
		html += `<li><div class="cart-item-img"><img src="` + *buildImagesHostname + passItem.Building.ImgUrl +
			`"></div><div class="cart-item-desc"><div>Услуга: ` + passItem.Building.Name +
			`</div><div>Описание: ` + passItem.Building.Description +
			`</div></div><div class="cart-item-quantity">Комментарий: ` + passItem.Comment +
			`</div></li>` + "\n"
	}

	return &html
}
