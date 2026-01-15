import request from './request'

// 获取商品列表
export function getProductList(params) {
  return request.get('/product/products', { params })
}

// 获取商品详情
export function getProductDetail(id) {
  return request.get(`/product/products/${id}`)
}

// 搜索商品
export function searchProducts(keyword, params) {
  return request.get('/product/search', {
    params: {
      keyword,
      ...params
    }
  })
}

// 获取类目列表
export function getCategoryList() {
  return request.get('/product/categories')
}

// 获取品牌列表
export function getBrandList(params) {
  return request.get('/product/brands', { params })
}

