<template>
  <el-footer class="app-footer">
    <div class="footer-content">
      <div class="footer-section">
        <h4>关于我们</h4>
        <ul>
          <li><a href="#">公司简介</a></li>
          <li><a href="#">联系我们</a></li>
          <li><a href="#">招聘信息</a></li>
        </ul>
      </div>
      
      <div class="footer-section">
        <h4>购物指南</h4>
        <ul>
          <li><a href="#">购物流程</a></li>
          <li><a href="#">会员介绍</a></li>
          <li><a href="#">常见问题</a></li>
        </ul>
      </div>
      
      <div class="footer-section">
        <h4>售后服务</h4>
        <ul>
          <li><a href="#">退换货政策</a></li>
          <li><a href="#">退款说明</a></li>
          <li><a href="#">联系客服</a></li>
        </ul>
      </div>
      
      <div class="footer-section">
        <h4>关注我们</h4>
        <ul>
          <li><a href="#">微信公众号</a></li>
          <li><a href="#">官方微博</a></li>
          <li><a href="#">抖音号</a></li>
        </ul>
      </div>
    </div>
    
    <div class="footer-bottom">
      <p>&copy; 2024 zjMall. All rights reserved.</p>
    </div>
  </el-footer>
</template>

<script setup>
</script>

<style scoped>
.app-footer {
  background: linear-gradient(135deg, #1e3c72 0%, #2a5298 100%);
  color: #fff;
  padding: 60px 20px 30px;
  margin-top: auto;
  width: 100%;
  box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.1);
  position: relative;
  overflow: hidden;
}

.app-footer::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, #409eff, #67c23a, #e6a23c, #f56c6c);
}

.footer-content {
  max-width: 1200px;
  margin: 0 auto;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 50px;
  margin-bottom: 40px;
  position: relative;
  z-index: 1;
}

.footer-section {
  transition: transform 0.3s ease;
}

.footer-section:hover {
  transform: translateY(-5px);
}

.footer-section h4 {
  margin-bottom: 25px;
  font-size: 18px;
  font-weight: 600;
  color: #fff;
  position: relative;
  padding-bottom: 12px;
  letter-spacing: 0.5px;
}

.footer-section h4::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  width: 40px;
  height: 3px;
  background: linear-gradient(90deg, #409eff, transparent);
  border-radius: 2px;
}

.footer-section ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.footer-section ul li {
  margin-bottom: 12px;
  position: relative;
  padding-left: 15px;
}

.footer-section ul li::before {
  content: '•';
  position: absolute;
  left: 0;
  color: #409eff;
  font-weight: bold;
  font-size: 16px;
}

.footer-section ul li a {
  color: #e8e8e8;
  text-decoration: none;
  transition: all 0.3s ease;
  font-size: 14px;
  display: inline-block;
  line-height: 1.6;
}

.footer-section ul li a:hover {
  color: #409eff;
  transform: translateX(5px);
  padding-left: 5px;
}

.footer-bottom {
  text-align: center;
  padding-top: 30px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  color: #b8c5d1;
  font-size: 14px;
  position: relative;
  z-index: 1;
}

.footer-bottom p {
  margin: 0;
  letter-spacing: 0.5px;
}

/* 响应式设计 */
@media (max-width: 992px) {
  .footer-content {
    grid-template-columns: repeat(2, 1fr);
    gap: 40px;
  }
}

@media (max-width: 576px) {
  .footer-content {
    grid-template-columns: 1fr;
    gap: 30px;
  }
  
  .app-footer {
    padding: 40px 20px 20px;
  }
}
</style>

