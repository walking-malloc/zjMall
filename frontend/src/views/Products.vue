<template>
  <div class="products-page">
        <div class="products-container">
          <div class="filter-section">
            <el-row :gutter="20">
              <el-col :span="6">
                <el-select 
                  v-model="selectedCategory" 
                  placeholder="选择类目" 
                  clearable 
                  @change="handleFilterChange"
                  teleported
                  :popper-options="{
                    modifiers: [
                      {
                        name: 'offset',
                        options: { offset: [0, 4] }
                      },
                      {
                        name: 'flip',
                        enabled: false
                      }
                    ],
                    placement: 'bottom-start'
                  }"
                >
                  <el-option label="全部" value="" />
                  <el-option
                    v-for="category in categories"
                    :key="category.id"
                    :label="category.name"
                    :value="category.id"
                  />
                </el-select>
              </el-col>
              <el-col :span="6">
                <el-select 
                  v-model="selectedBrand" 
                  placeholder="选择品牌" 
                  clearable 
                  @change="handleFilterChange"
                  teleported
                  :popper-options="{
                    modifiers: [
                      {
                        name: 'offset',
                        options: { offset: [0, 4] }
                      },
                      {
                        name: 'flip',
                        enabled: false
                      }
                    ],
                    placement: 'bottom-start'
                  }"
                >
                  <el-option label="全部" value="" />
                  <el-option
                    v-for="brand in brands"
                    :key="brand.id"
                    :label="brand.name"
                    :value="brand.id"
                  />
                </el-select>
              </el-col>
              <el-col :span="6">
                <el-select 
                  v-model="sortBy" 
                  placeholder="排序方式" 
                  @change="handleFilterChange"
                  teleported
                  :popper-options="{
                    modifiers: [
                      {
                        name: 'offset',
                        options: { offset: [0, 4] }
                      },
                      {
                        name: 'flip',
                        enabled: false
                      }
                    ],
                    placement: 'bottom-start'
                  }"
                >
                  <el-option label="默认排序" value="default" />
                  <el-option label="价格从低到高" value="price_asc" />
                  <el-option label="价格从高到低" value="price_desc" />
                  <el-option label="最新上架" value="newest" />
                </el-select>
              </el-col>
            </el-row>
          </div>

          <div class="products-list" v-loading="loading">
            <el-row :gutter="20">
              <el-col :span="6" v-for="product in products" :key="product.id">
                <el-card class="product-card" @click="goToDetail(product.id)">
                  <img :src="product.main_image || '/placeholder.png'" class="product-image" />
                  <div class="product-info">
                    <h4>{{ product.title }}</h4>
                    <p class="subtitle">{{ product.subtitle }}</p>
                    <p class="price">¥{{ product.price || '0.00' }}</p>
                  </div>
                </el-card>
              </el-col>
            </el-row>

            <el-empty v-if="!loading && products.length === 0" description="暂无商品" />

            <el-pagination
              v-if="total > 0"
              v-model:current-page="currentPage"
              v-model:page-size="pageSize"
              :total="total"
              :page-sizes="[12, 24, 48]"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="handleSizeChange"
              @current-change="handlePageChange"
              style="margin-top: 20px; justify-content: center;"
            />
          </div>
        </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { getProductList, getCategoryList, searchProducts, getBrandList } from '@/api/product'

const router = useRouter()
const route = useRoute()

const searchKeyword = ref(route.query.keyword || '')
const selectedCategory = ref(route.query.category_id || '')
const selectedBrand = ref('')
const sortBy = ref('default')
const categories = ref([])
const brands = ref([])
const products = ref([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(12)
const total = ref(0)

const loadCategories = async () => {
  try {
    const res = await getCategoryList()
    if (res.data.code === 0) {
      categories.value = res.data.data || []
    }
  } catch (error) {
    console.error('获取类目失败:', error)
  }
}

const loadBrands = async () => {
  try {
    const res = await getBrandList({ page: 1, page_size: 100 })
    if (res.data.code === 0) {
      brands.value = res.data.data || []
    }
  } catch (error) {
    console.error('获取品牌失败:', error)
  }
}

const loadProducts = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      page_size: pageSize.value
    }
    
    if (selectedCategory.value) {
      params.category_id = selectedCategory.value
    }
    
    if (selectedBrand.value) {
      params.brand_id = selectedBrand.value
    }

    let res
    if (searchKeyword.value) {
      res = await searchProducts(searchKeyword.value, params)
      console.log('搜索商品响应:', res.data)
      // SearchProductsResponse 返回的字段是 products
      if (res.data.code === 0) {
        products.value = res.data.products || []
        total.value = res.data.total || 0
      } else {
        console.error('搜索商品失败:', res.data.message)
      }
    } else {
      res = await getProductList(params)
      console.log('商品列表响应:', res.data)
      // ListProductsResponse 返回的字段是 data
      if (res.data.code === 0) {
        products.value = res.data.data || []
        total.value = res.data.total || 0
        console.log('商品数据:', products.value)
      } else {
        console.error('获取商品列表失败:', res.data.message)
      }
    }
  } catch (error) {
    console.error('获取商品列表失败:', error)
    console.error('错误详情:', error.response?.data || error.message)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  currentPage.value = 1
  loadProducts()
}

const handleFilterChange = () => {
  currentPage.value = 1
  loadProducts()
}

const handleSizeChange = () => {
  currentPage.value = 1
  loadProducts()
}

const handlePageChange = () => {
  loadProducts()
}

const goToDetail = (id) => {
  router.push(`/product/products/${id}`)
}

watch(selectedCategory, () => {
  currentPage.value = 1
  loadProducts()
})

onMounted(() => {
  loadCategories()
  loadBrands()
  loadProducts()
})
</script>

<style scoped>
.products-page {
  min-height: calc(100vh - 140px);
  padding: 20px;
  background: #f5f5f5;
}

.products-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.filter-section {
  margin-bottom: 20px;
}

.products-list {
  min-height: 400px;
}

.product-card {
  cursor: pointer;
  transition: transform 0.3s;
  margin-bottom: 20px;
}

.product-card:hover {
  transform: translateY(-5px);
}

.product-image {
  width: 100%;
  height: 200px;
  object-fit: cover;
}

.product-info {
  padding: 10px 0;
}

.product-info h4 {
  font-size: 16px;
  margin-bottom: 5px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.subtitle {
  font-size: 12px;
  color: #999;
  margin-bottom: 10px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.price {
  color: #f56c6c;
  font-size: 20px;
  font-weight: bold;
}
</style>

<style>
/* 全局样式：确保下拉框在下方显示，防止自动调整到上方 */
.el-select-dropdown {
  margin-top: 0 !important;
}

/* 强制下拉框在下方显示 */
.el-popper[data-popper-placement^="top"] {
  transform: translateY(0) !important;
  margin-top: 0 !important;
}

/* 确保下拉框始终在下方 */
.el-select-dropdown__wrap {
  max-height: 274px;
}
</style>

