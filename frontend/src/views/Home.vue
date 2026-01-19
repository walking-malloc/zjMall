<template>
  <div class="home">
    <div class="banner">
      <h2>欢迎来到 zjMall</h2>
      <p>发现更多优质商品</p>
      <el-button type="primary" size="large" @click="$router.push('/product/products')">
        开始购物
      </el-button>
    </div>

    <!-- 分类导航 -->
    <div class="category-section">
      <h3>商品分类</h3>
      <el-row :gutter="10">
        <el-col :span="3" v-for="category in categories" :key="category.id"
          @click="goToCategory(category.id)">
          <div class="category-item">
            <div class="category-icon">
              <el-icon :size="40">
                <Box />
              </el-icon>
            </div>
            <p>{{ category.name }}</p>
          </div>
        </el-col>
      </el-row>
      <el-empty v-if="!loading && categories.length === 0" description="暂无分类" />
    </div>

    <!-- 热门商品 -->
    <div class="hot-products">
      <h3>热门商品</h3>
      <el-empty v-if="!loading && products.length === 0" description="暂无商品" />
      <el-row :gutter="20" v-loading="loading" v-else>
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
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getProductList, getCategoryList } from '@/api/product'
import { Box } from '@element-plus/icons-vue'

const router = useRouter()

const products = ref([])
const categories = ref([])
const loading = ref(false)

const goToDetail = (id) => {
  router.push(`/product/products/${id}`)
}


onMounted(async () => {
  loading.value = true
  try {
    const [productsRes, categoriesRes] = await Promise.all([
      getProductList({ page: 1, page_size: 8 }),
      // 使用 ListCategories 接口，只获取一级分类且可见的分类
      getCategoryList({ 
        is_visible: true // 前台可见
      })
    ])

    console.log('商品列表响应:', productsRes.data)
    console.log('分类列表响应:', categoriesRes.data)

    if (productsRes.data.code === 0) {
      products.value = productsRes.data.data || []
      console.log('商品数据:', products.value)
    } else {
      console.error('获取商品列表失败:', productsRes.data.message)
    }

    if (categoriesRes.data.code === 0) {
      // ListCategories 接口返回的 Data 是数组
      const categoryData = categoriesRes.data.data
      if (Array.isArray(categoryData)) {
        categories.value = categoryData
      } else {
        categories.value = []
      }
      console.log('分类数据:', categories.value)
      console.log('分类总数:', categoriesRes.data.total || categories.value.length)
    } else {
      console.error('获取分类列表失败:', categoriesRes.data.message)
      categories.value = []
    }
  } catch (error) {
    console.error('获取数据失败:', error)
    console.error('错误详情:', error.response?.data || error.message)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.home {
  min-height: 100vh;
}

.home {
  padding: 20px;
}

.banner {
  text-align: center;
  padding: 60px 0;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  margin-bottom: 40px;
}

.banner h2 {
  font-size: 48px;
  margin-bottom: 20px;
}

.banner p {
  font-size: 20px;
  margin-bottom: 30px;
}

.hot-products {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
  min-height: 300px;
}

.category-section {
  max-width: 1200px;
  margin: 0 auto 40px;
  padding: 0 20px;
}

.category-section h3 {
  font-size: 24px;
  margin-bottom: 20px;
}

.category-item {
  text-align: center;
  padding: 20px;
  background: #fff;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
  border: 1px solid #eee;
}

.category-item:hover {
  transform: translateY(-5px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.category-icon {
  color: #409eff;
  margin-bottom: 10px;
}

.category-item p {
  margin: 0;
  font-size: 14px;
  color: #333;
}

.hot-products h3 {
  font-size: 24px;
  margin-bottom: 20px;
}

.subtitle {
  font-size: 12px;
  color: #999;
  margin: 5px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-card {
  cursor: pointer;
  transition: transform 0.3s;
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
