package utils

type Pager struct {
	PageNo   int    `form:"pageno"`
	PageSize int    `form:"pagesize"`
	SortBy   string `form:"sortBy"`
	SortDesc string `form:"sortDesc"`
	Sort     string `form:"sort"` //新的从前端穿过来的排序（多列支持）
}

func (form *Pager) Paging() (int, int, int, string) {
	pageno := form.PageNo
	if pageno == 0 {
		pageno = 1
	} // default value
	pagesize := form.PageSize
	if pagesize == 0 {
		pagesize = 10
	} // default value
	var sort string
	if form.Sort != "" {
		sort = form.Sort
	} else {
		if form.SortBy == "" {
			sort = "id"
		} else {
			sort = form.SortBy
		}
		if form.SortDesc == "" {
			sort += " desc"
		} else {
			sort += " " + form.SortDesc
		}
	}
	start := (pageno - 1) * pagesize
	return pageno, pagesize, start, sort
}
