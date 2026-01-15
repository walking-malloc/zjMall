<template>
  <div class="product-detail-page">
    <div class="detail-container" v-loading="loading">
      <el-row :gutter="40" v-if="product">
        <el-col :span="12">
          <div class="image-section">
            <el-image
              :src="currentImage || product.main_image || '/placeholder.png'"
              fit="contain"
              class="main-image"
            />
            <div class="image-thumbs" v-if="productImages.length > 1">
              <div
                v-for="(img, index) in productImages"
                :key="index"
                class="thumb-item"
                :class="{ active: currentImage === img }"
                @click="currentImage = img"
              >
                <img :src="img" />
              </div>
            </div>
          </div>
        </el-col>
        <el-col :span="12">
          <h1>{{ product.title }}</h1>
          <p class="subtitle">{{ product.subtitle }}</p>
          <div class="price-section">
            <span class="price">¥{{ selectedSku?.price || product.price || '0.00' }}</span>
          </div>
          
          <!-- SKU 选择 -->
          <div class="sku-section" v-if="skus && skus.length > 0">
            <div v-for="(sku, index) in skus" :key="index" class="sku-item">
              <span class="sku-label">{{ sku.attribute_name }}：</span>
              <el-radio-group v-model="selectedSkuId" @change="handleSkuChange">
                <el-radio 
                  v-for="value in sku.values" 
                  :key="value.id"
                  :label="value.id"
                >
                  {{ value.value }}
                </el-radio>
              </el-radio-group>
            </div>
          </div>

          <!-- 数量选择 -->
          <div class="quantity-section">
            <span class="label">数量：</span>
            <el-input-number
              v-model="quantity"
              :min="1"
              :max="selectedSku?.stock || 999"
            />
            <span class="stock-info" v-if="selectedSku">
              库存：{{ selectedSku.stock }}
            </span>
          </div>

          <div class="actions">
            <el-button type="primary" size="large" @click="handleAddToCart">
              加入购物车
            </el-button>
            <el-button type="danger" size="large" @click="handleBuyNow">
              立即购买
            </el-button>
          </div>

          <!-- 商品详情 -->
          <el-tabs v-model="activeTab" style="margin-top: 40px;">
            <el-tab-pane label="商品详情" name="detail">
              <div class="product-description" v-html="product.description || '暂无详情'"></div>
            </el-tab-pane>
            <el-tab-pane label="规格参数" name="spec">
              <el-descriptions :column="2" border>
                <el-descriptions-item label="商品ID">{{ product.id }}</el-descriptions-item>
                <el-descriptions-item label="品牌">{{ product.brand_name || '暂无' }}</el-descriptions-item>
                <el-descriptions-item label="类目">{{ product.category_name || '暂无' }}</el-descriptions-item>
                <el-descriptions-item label="状态">
                  <el-tag :type="product.status === 4 ? 'success' : 'info'">
                    {{ product.status === 4 ? '已上架' : '未上架' }}
                  </el-tag>
                </el-descriptions-item>
              </el-descriptions>
            </el-tab-pane>
          </el-tabs>
        </el-col>
      </el-row>
      
      <el-empty v-if="!loading && !product" description="商品不存在" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getProductDetail } from '@/api/product'
import { useCartStore } from '@/stores/cart'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const cartStore = useCartStore()

const product = ref(null)
const skus = ref([])
const loading = ref(false)
const quantity = ref(1)
const selectedSkuId = ref(null)
const activeTab = ref('detail')
const currentImage = ref('')

const productImages = computed(() => {
  if (!product.value) return []
  const images = []
  if (product.value.main_image) {
    images.push(product.value.main_image)
  }
  // TODO: 解析 product.images JSON 字符串
  return images
})

const selectedSku = computed(() => {
  if (!selectedSkuId.value || !skus.value.length) return null
  // TODO: 根据 selectedSkuId 找到对应的 SKU
  return null
})

const handleSkuChange = () => {
  quantity.value = 1
}

const loadProduct = async () => {
  loading.value = true
  try {
    const res = await getProductDetail(route.params.id)
    if (res.data.code === 0) {
      product.value = res.data.product
      skus.value = res.data.skus || []
      currentImage.value = product.value.main_image || ''
      
      // 如果有 SKU，默认选择第一个
      if (skus.value.length > 0 && skus.value[0].values?.length > 0) {
        selectedSkuId.value = skus.value[0].values[0].id
      }
    } else {
      ElMessage.error('获取商品详情失败')
    }
  } catch (error) {
    ElMessage.error('获取商品详情失败')
  } finally {
    loading.value = false
  }
}

const handleAddToCart = () => {
  if (!product.value) return
  
  cartStore.addItem({
    id: product.value.id,
    title: product.value.title,
    main_image: product.value.main_image,
    price: selectedSku.value?.price || product.value.price,
    skuId: selectedSkuId.value,
    quantity: quantity.value
  })
  
  ElMessage.success('已添加到购物车')
}

const handleBuyNow = () => {
  if (!product.value) return
  
  // 先添加到购物车
  handleAddToCart()
  
  // 跳转到结算页面
  router.push({
    path: '/checkout',
    query: {
      items: JSON.stringify([{
        productId: product.value.id,
        skuId: selectedSkuId.value,
        quantity: quantity.value
      }])
    }
  })
}

onMounted(() => {
  loadProduct()
})
</script>

<style scoped>
.product-detail-page {
  min-height: calc(100vh - 140px);
  padding: 20px;
  background: #f5f5f5;
}

.detail-container {
  max-width: 1200px;
  margin: 0 auto;
  background: #fff;
  padding: 40px;
  border-radius: 4px;
}

.image-section {
  position: sticky;
  top: 20px;
}

.main-image {
  width: 100%;
  height: 500px;
  border: 1px solid #eee;
  border-radius: 4px;
}

.image-thumbs {
  display: flex;
  gap: 10px;
  margin-top: 10px;
}

.thumb-item {
  width: 80px;
  height: 80px;
  border: 2px solid transparent;
  border-radius: 4px;
  cursor: pointer;
  overflow: hidden;
}

.thumb-item.active {
  border-color: #409eff;
}

.thumb-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.product-detail-page h1 {
  font-size: 24px;
  margin-bottom: 10px;
}

.subtitle {
  color: #999;
  font-size: 16px;
  margin: 20px 0;
}

.price-section {
  margin: 30px 0;
  padding: 20px;
  background: #f5f5f5;
  border-radius: 4px;
}

.price {
  color: #f56c6c;
  font-size: 36px;
  font-weight: bold;
}

.sku-section {
  margin: 30px 0;
}

.sku-item {
  margin-bottom: 20px;
}

.sku-label {
  display: inline-block;
  width: 80px;
  font-weight: bold;
}

.quantity-section {
  margin: 30px 0;
  display: flex;
  align-items: center;
  gap: 15px;
}

.quantity-section .label {
  font-weight: bold;
}

.stock-info {
  color: #999;
  font-size: 14px;
}

.actions {
  margin-top: 40px;
  display: flex;
  gap: 20px;
}

.product-description {
  line-height: 1.8;
  color: #666;
}
</style>

