// Only official metadata labels use this table. Never pass free-text draft values here.
const commonLabels: Record<string, string> = {
  Brand: "品牌", Marca: "品牌", Model: "型号", Modelo: "型号",
  "Universal product code": "商品通用编码", "Código universal de producto": "商品通用编码",
  "Código universal do produto": "商品通用编码",
  "Item condition": "商品成色", "Condición del ítem": "商品成色", "Condição do item": "商品成色",
  "Sale format": "销售形式", "Formato de venta": "销售形式", "Formato de venda": "销售形式",
  "Transportabilidad Hazmat": "危险品运输属性", "Transportabilidade Hazmat": "危险品运输属性",
  "Cantidad de eliminadores de olor para electrodomésticos": "家电除臭剂数量",
  Other: "其他属性", Others: "其他属性", Otros: "其他属性", Outros: "其他属性",
  Color: "颜色", Cor: "颜色", Size: "尺寸", Talla: "尺码", Tamanho: "尺寸",
  Material: "材质", New: "全新", Nuevo: "全新", Novo: "全新",
  Used: "二手", Usado: "二手", Unit: "单件", Unidad: "单件", Unidade: "单件",
  Exceptuado: "豁免", Isento: "豁免",
};
const identifierFields = /^(BRAND|MODEL|SELLER_SKU|GTIN|EAN|UPC|ISBN|MPN|PART_NUMBER|SERIAL_NUMBER)$/;

export function publishedMetadataLabel(name: string, translations: Record<string, string>): string {
  const translated = Object.hasOwn(translations, name) ? translations[name]?.trim() : "";
  if (translated && translated !== name) return translated;
  return Object.hasOwn(commonLabels, name) ? commonLabels[name] : name;
}

export function publishedChoiceLabel(fieldId: string, name: string, translations: Record<string, string>): string {
  return identifierFields.test(fieldId) ? name : publishedMetadataLabel(name, translations);
}
