package entity

type Person struct {
    ID          int     `json:"id"`
    Name        string  `json:"name"`
    Surname     string  `json:"surname"`
    Patronymic  *string `json:"patronymic,omitempty"`
    Age         int     `json:"age"`
    Gender      string  `json:"gender"`
    Nationality string  `json:"nationality"`
}

type PersonInput struct {
    Name       string  `json:"name" binding:"required"`
    Surname    string  `json:"surname" binding:"required"`
    Patronymic *string `json:"patronymic,omitempty"`
}

type PersonFilter struct {
    Name        *string `form:"name"`
    Surname     *string `form:"surname"`
    Patronymic  *string `form:"patronymic"`
    AgeMin      *int    `form:"age_min"`
    AgeMax      *int    `form:"age_max"`
    Gender      *string `form:"gender"`
    Nationality *string `form:"nationality"`
    Page        int     `form:"page,default=1"`
    PageSize    int     `form:"page_size,default=10"`
}