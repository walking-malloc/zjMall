import request from './request'

// 获取商品列表
export function getProductList(params) {
  return request.get('/product/products', { params })
}

// 获取商品详情
export function getProductDetail(id) {
  return request.get(`/product/products/${id}`, {
    params: {
      include_skus: true,
      include_tags: false
    }
  })
}

// 获取单个 SKU 详情（包含规格属性）
export function getSkuDetail(skuId) {
  return request.get(`/product/skus/${skuId}`, {
    params: {
      include_attributes: true
    }
  })
}

// 搜索商品
export function searchProducts(keyword, params) {
  return request.get('/product/search', {
    params: {
      keyword,//表示这个参数是必传的
      ...params
    }
  })
}

// 获取类目列表（使用 ListCategories 接口）
export function getCategoryList(params = {}) {
  return request.get('/product/categories', {
    params: {
      status: 1,
      is_visible: true,
      ...params
    }
  })
}

// 获取品牌列表
export function getBrandList(params) {
  return request.get('/product/brands', { params })
}

