using System;
using Enumer;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using Newtonsoft.Json;
using ZYing.Data;
using ZYing.Meli;

namespace ZYing.Channel;

public class Meliup
{
	private MeliUpVariants _variants;

	[Column("up_aid", PBType.Integer, Primary = true, NotNull = true, Memo = "站点id")]
	public int Aid { get; set; }

	[Column("up_saleid", PBType.Bigint, Primary = true, NotNull = true, Memo = "产品id")]
	public long Saleid { get; set; }

	[Column("up_fromid", PBType.Integer, NotNull = true, Memo = "来源 userid")]
	public int? Fromid { get; set; }

	[Column("up_group", PBType.Integer, NotNull = true, Memo = "账号groupid")]
	public int? Group { get; set; }

	[Column("up_siteid", PBType.Varchar, 3, NotNull = true, Memo = "站点site id")]
	public string Siteid { get; set; }

	[Column("up_loginid", PBType.Integer, NotNull = true, Memo = "刊登业务员id")]
	public int? Loginid { get; set; }

	[Column("up_localid", PBType.Integer, NotNull = true, Memo = "分类id")]
	public int? Localid { get; set; }

	[Column("up_cateid", PBType.Integer, NotNull = true, Memo = "美客多cbt分类编号")]
	public int? Cateid { get; set; }

	[Column("up_sitecateid", PBType.Varchar, 50, Memo = "对应市场分类编号")]
	public string Sitecateid { get; set; }

	[Column("up_thumb", PBType.Varchar, 500, NotNull = true, Memo = "首图")]
	public string Thumb { get; set; }

	[Column("up_pics", (PBType)(-2147483626), Memo = "产品图片")]
	public string[] Pics { get; set; }

	[Column("up_familyid", PBType.Bigint, NotNull = true, Memo = "family_id")]
	public long? Familyid { get; set; }

	[Column("up_familyname", PBType.Varchar, 1000, NotNull = true, Memo = "family_name")]
	public string Familyname { get; set; }

	[Column("up_globalid", PBType.Varchar, 50, Memo = "global_id")]
	public string Globalid { get; set; }

	[Column("up_upid", PBType.Varchar, 50, Memo = "user product id")]
	public string Upid { get; set; }

	[Column("up_cbtid", PBType.Varchar, 50, Memo = "cbtId")]
	public string Cbtid { get; set; }

	[Column("up_itemid", PBType.Varchar, 50, Memo = "itemid")]
	public string Itemid { get; set; }

	[Column("up_productid", PBType.Varchar, 255, Memo = "跟卖目录id")]
	public string Productid { get; set; }

	[Column("up_variations", PBType.Json, Memo = "变体json")]
	public string Variations { get; set; }

	[Column("up_stat", PBType.Integer, NotNull = true, Default = "0", Memo = "上传状态")]
	public int? Stat { get; set; }

	[Column("up_status", PBType.Integer, NotNull = true, Default = "0", Memo = "线上状态")]
	public int? Status { get; set; }

	[Column("up_link", PBType.Varchar, 500, Memo = "前台页link")]
	public string Link { get; set; }

	[Column("up_title", PBType.Varchar, 500, NotNull = true, Memo = "产品标题")]
	public string Title { get; set; }

	[Column("up_intro", PBType.Text, Memo = "产品描述")]
	public string Intro { get; set; }

	[Column("up_note", PBType.Text, Memo = "调价说明")]
	public string Note { get; set; }

	[Column("up_reason", PBType.Text, Memo = "错误原因")]
	public string Reason { get; set; }

	[Column("up_attrs", PBType.Json, Memo = "产品属性")]
	public string Attrs { get; set; }

	[Column("up_weight", PBType.Integer, NotNull = true, Default = "0", Memo = "产品重量")]
	public int? Weight { get; set; }

	[Column("up_size", (PBType)(-2147483639), Memo = "产品尺寸")]
	public int[] Size { get; set; }

	[Column("up_stock", PBType.Integer, NotNull = true, Default = "0", Memo = "总库存数")]
	public int? Stock { get; set; }

	[Column("up_price", PBType.Numeric, NotNull = true, Default = "0", Memo = "前台售价")]
	public decimal? Price { get; set; }

	[Column("up_price_discount", PBType.Numeric, NotNull = true, Default = "0", Memo = "折后价")]
	public decimal? Price_discount { get; set; }

	[Column("up_net_proceeds", PBType.Numeric, NotNull = true, Default = "0", Memo = "净收益价")]
	public decimal? Net_proceeds { get; set; }

	[Column("up_shipcost", PBType.Numeric, NotNull = true, Default = "0", Memo = "平台运费")]
	public decimal? Shipcost { get; set; }

	[Column("up_salefee", PBType.Numeric, NotNull = true, Default = "0", Memo = "平台佣金")]
	public decimal? Salefee { get; set; }

	[Column("up_sellprice", PBType.Numeric, NotNull = true, Default = "0", Memo = "本币价格")]
	public decimal? Sellprice { get; set; }

	[Column("up_warnning", PBType.Varchar, 500, Memo = "平台警告信息")]
	public string Warnning { get; set; }

	[Column("up_orders", PBType.Integer, NotNull = true, Default = "0", Memo = "订单数量")]
	public int? Orders { get; set; }

	[Column("up_view", PBType.Integer, NotNull = true, Default = "0", Memo = "浏览量")]
	public int? View { get; set; }

	[Column("up_health", PBType.Integer, NotNull = true, Default = "0", Memo = "产品评分")]
	public int? Health { get; set; }

	[Column("up_suggest", (PBType)(-2147483626), Memo = "健康分数建议")]
	public string[] Suggest { get; set; }

	[Column("up_ids", (PBType)(-2147483626), 50, Memo = "存放所有变体id，搜索用")]
	public string[] Ids { get; set; }

	[Column("up_synctime", PBType.Timestamp, Memo = "同步时间")]
	public DateTime? Synctime { get; set; }

	[Column("up_visittime", PBType.Timestamp, Memo = "访问量同步时间")]
	public DateTime? Visittime { get; set; }

	[Column("up_pubtime", PBType.Timestamp, Memo = "发布时间")]
	public DateTime? Pubtime { get; set; }

	public MeliUpVariants Variants
	{
		get
		{
			if (_variants == null)
			{
				_variants = ((!Variations.IsNullOrEmpty()) ? JsonConvert.DeserializeObject<MeliUpVariants>(Variations) : new MeliUpVariants());
			}
			return _variants;
		}
	}

	public string StatusText => Enumer.Meli.GetStatV2(Status.GetValueOrDefault())?.Text;

	public bool? Successed { get; set; }

	public string Messge { get; set; }

	public string ErrorDetail { get; set; }

	public Meliup()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.fmiGD6eDN();
		base..ctor();
	}
}
