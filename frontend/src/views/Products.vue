<template>
  <div class="products-page">
    <el-container>
      <el-header>
        <div class="header-content">
          <div class="logo" @click="$router.push('/')">
            <h1>zjMall</h1>
          </div>
          <div class="search-bar">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索商品"
              @keyup.enter="handleSearch"
            >
              <template #append>
                <el-button @click="handleSearch">搜索</el-button>
              </template>
            </el-input>
          </div>
          <div class="user-actions">
            <el-button type="text" @click="$router.push('/')">首页</el-button>
            <el-button v-if="!userStore.isLoggedIn" type="text" @click="$router.push('/login')">登录</el-button>
          </div>
        </div>
      </el-header>

      <el-main>
        <div class="products-container">
          <div class="filter-section">
            <el-select v-model="selectedCategory" placeholder="选择类目" clearable>
              <el-option label="全部" value="" />
              <el-option
                v-for="category in categories"
                :key="category.id"
                :label="category.name"
                :value="category.id"
              />
            </el-select>
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
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { getProductList, getCategoryList, searchProducts } from '@/api/product'

const router = useRouter()
const userStore = useUserStore()

const searchKeyword = ref('')
const selectedCategory = ref('')
const categories = ref([])
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

    let res
    if (searchKeyword.value) {
      res = await searchProducts(searchKeyword.value, params)
    } else {
      res = await getProductList(params)
    }

    if (res.data.code === 0) {
      products.value = res.data.data || []
      total.value = res.data.total || 0
    }
  } catch (error) {
    console.error('获取商品列表失败:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
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
  router.push(`/product/${id}`)
}

watch(selectedCategory, () => {
  currentPage.value = 1
  loadProducts()
})

onMounted(() => {
  loadCategories()
  loadProducts()
})
</script>

<style scoped>
.products-page {
  min-height: 100vh;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.logo h1 {
  margin: 0;
  color: #409eff;
  cursor: pointer;
}

.search-bar {
  flex: 1;
  max-width: 500px;
  margin: 0 20px;
}

.user-actions {
  display: flex;
  gap: 10px;
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

