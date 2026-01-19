-- 插入测试品牌数据
USE product_db;

-- 插入一些常见品牌用于测试
INSERT INTO brands (id, name, logo_url, country, description, first_letter, sort_order, status, version) VALUES
-- 电子产品品牌
('01KF0000000000000000000001', 'Apple', 'https://logo.clearbit.com/apple.com', '美国', '苹果公司，全球知名科技品牌', 'A', 100, 1, 0),
('01KF0000000000000000000002', 'Samsung', 'https://logo.clearbit.com/samsung.com', '韩国', '三星电子，全球领先的消费电子品牌', 'S', 99, 1, 0),
('01KF0000000000000000000003', 'Huawei', 'https://logo.clearbit.com/huawei.com', '中国', '华为技术有限公司，全球领先的信息与通信技术解决方案供应商', 'H', 98, 1, 0),
('01KF0000000000000000000004', 'Xiaomi', 'https://logo.clearbit.com/mi.com', '中国', '小米科技，以性价比著称的智能硬件品牌', 'X', 97, 1, 0),
('01KF0000000000000000000005', 'OPPO', 'https://logo.clearbit.com/oppo.com', '中国', 'OPPO，专注于拍照技术的手机品牌', 'O', 96, 1, 0),
('01KF0000000000000000000006', 'Vivo', 'https://logo.clearbit.com/vivo.com', '中国', 'vivo，专注于音乐和拍照的智能手机品牌', 'V', 95, 1, 0),
('01KF0000000000000000000007', 'Sony', 'https://logo.clearbit.com/sony.com', '日本', '索尼，全球知名的消费电子和娱乐公司', 'S', 94, 1, 0),
('01KF0000000000000000000008', 'Lenovo', 'https://logo.clearbit.com/lenovo.com', '中国', '联想，全球领先的PC和智能设备制造商', 'L', 93, 1, 0),

-- 服装品牌
('01KF0000000000000000000009', 'Nike', 'https://logo.clearbit.com/nike.com', '美国', '耐克，全球领先的运动品牌', 'N', 90, 1, 0),
('01KF0000000000000000000010', 'Adidas', 'https://logo.clearbit.com/adidas.com', '德国', '阿迪达斯，世界知名运动品牌', 'A', 89, 1, 0),
('01KF0000000000000000000011', 'Uniqlo', 'https://logo.clearbit.com/uniqlo.com', '日本', '优衣库，日本快时尚品牌', 'U', 88, 1, 0),
('01KF0000000000000000000012', 'ZARA', 'https://logo.clearbit.com/zara.com', '西班牙', 'ZARA，西班牙快时尚品牌', 'Z', 87, 1, 0),
('01KF0000000000000000000013', 'H&M', 'https://logo.clearbit.com/hm.com', '瑞典', 'H&M，瑞典快时尚品牌', 'H', 86, 1, 0),

-- 美妆品牌
('01KF0000000000000000000014', 'L\'Oreal', 'https://logo.clearbit.com/loreal.com', '法国', '欧莱雅，全球知名美妆品牌', 'L', 85, 1, 0),
('01KF0000000000000000000015', 'Estee Lauder', 'https://logo.clearbit.com/esteelauder.com', '美国', '雅诗兰黛，高端美妆品牌', 'E', 84, 1, 0),
('01KF0000000000000000000016', 'MAC', 'https://logo.clearbit.com/maccosmetics.com', '加拿大', 'MAC，专业彩妆品牌', 'M', 83, 1, 0),
('01KF0000000000000000000017', 'Maybelline', 'https://logo.clearbit.com/maybelline.com', '美国', '美宝莲，大众美妆品牌', 'M', 82, 1, 0),

-- 家电品牌
('01KF0000000000000000000018', 'Haier', 'https://logo.clearbit.com/haier.com', '中国', '海尔，全球领先的家电品牌', 'H', 80, 1, 0),
('01KF0000000000000000000019', 'Midea', 'https://logo.clearbit.com/midea.com', '中国', '美的，中国家电领导品牌', 'M', 79, 1, 0),
('01KF0000000000000000000020', 'Gree', 'https://logo.clearbit.com/gree.com', '中国', '格力，专业空调制造商', 'G', 78, 1, 0),
('01KF0000000000000000000021', 'Panasonic', 'https://logo.clearbit.com/panasonic.com', '日本', '松下，日本知名家电品牌', 'P', 77, 1, 0),
('01KF0000000000000000000022', 'LG', 'https://logo.clearbit.com/lg.com', '韩国', 'LG，韩国知名电子品牌', 'L', 76, 1, 0);

