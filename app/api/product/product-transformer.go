package product

import (
	"gofiber-starterkit/app/models"
)

func TransformProductListItem(p *models.Product) ProductListItemResponse {
	return ProductListItemResponse{
		ID:       p.ID,
		Name:     p.Name,
		Price:    p.Price,
		Status:   p.Status,
		ImageURL: p.ImageURL,
	}
}

func TransformProductDetail(p *models.Product) ProductDetailResponse {
	return ProductDetailResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		Status:      p.Status,
		ImageURL:    p.ImageURL,
	}
}

func TransformProductList(products []*models.Product) []ProductListItemResponse {
	var transformed []ProductListItemResponse
	for _, p := range products {
		transformed = append(transformed, TransformProductListItem(p))
	}
	return transformed
}
