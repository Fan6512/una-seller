using System;
using System.Collections.Generic;
using System.Collections.Specialized;
using System.Data;
using System.Data.Common;
using System.Diagnostics;
using System.IO;
using System.Runtime.Serialization;
using System.Text;
using System.Text.RegularExpressions;
using System.Web;
using System.Xml;
using System.Xml.Schema;
using System.Xml.Serialization;
using ZYing.Web;

namespace ZYing.Data;

[Serializable]
[DebuggerDisplay("Record: {DebuggerDisplay,nq}")]
[DebuggerTypeProxy(typeof(RecordDebugView))]
public class Record : ISerializable, IXmlSerializable
{
	private class RecordDebugView
	{
		[DebuggerBrowsable(DebuggerBrowsableState.Never)]
		private readonly Record _record;

		public string Name
		{
			get
			{
				if (!string.IsNullOrWhiteSpace(_record.Name))
				{
					return _record.Name;
				}
				return "None";
			}
		}

		public int Count => _record.Count;

		public int MaxCount => _record.MaxCount;

		public int Page => _record.Page;

		public int PageSize => _record.PageSize;

		public string Data => _record.ToString();

		public RecordDebugView(Record record)
		{
			_record = record ?? throw new ArgumentNullException("record");
		}
	}

	public static int MAXLIMIT = 200000;

	public const int FixSize = 60;

	private int _count;

	[NonSerialized]
	private int _cursor;

	public string Name = "";

	public object Tag;

	public int Page = 1;

	private int _maxcount;

	private int _pagesize;

	public int Scale = 18;

	public int PageLimit;

	private RecordColumns _cols;

	[NonSerialized]
	private OrderBy _reading;

	public int Count
	{
		get
		{
			return _count;
		}
		private set
		{
			_count = value;
			if (MAXLIMIT > 0 && value >= MAXLIMIT)
			{
				throw new Exception("Record 数据大于 " + MAXLIMIT);
			}
		}
	}

	public int Cursor
	{
		get
		{
			return _cursor;
		}
		set
		{
			switch (_reading)
			{
			case OrderBy.ASC:
				if (value < _cursor)
				{
					throw new Exception("在Record.Read()期间，Cursor只能递增！");
				}
				break;
			case OrderBy.DESC:
				if (value > _cursor)
				{
					throw new Exception("在Record.ReadBack()期间，Cursor只能递减！");
				}
				break;
			}
			_cursor = value;
		}
	}

	public int MaxCount
	{
		get
		{
			if (_maxcount <= 0)
			{
				return Count;
			}
			return _maxcount;
		}
		set
		{
			_maxcount = value;
		}
	}

	public int PageSize
	{
		get
		{
			return _pagesize;
		}
		set
		{
			int pagesize = (RowSize = value);
			_pagesize = pagesize;
		}
	}

	public int MaxPage
	{
		get
		{
			if (_maxcount == 0)
			{
				return 0;
			}
			int num = ((_pagesize == 0) ? 20 : _pagesize);
			return (_maxcount - 1) / num + 1;
		}
	}

	public string Scope => (Page - 1) * PageSize + 1 + " - " + ((Page - 1) * PageSize + Count);

	public RecordColumns Columns => _cols;

	public RecordColumn this[string name] => Columns[name];

	public RecordColumn this[string name, int i] => Columns[name, i];

	public RecordColumn this[int col] => Columns[col];

	public RecordColumn this[int col, int i] => Columns[col, i];

	public int RowSize
	{
		get
		{
			return Columns.RowSize;
		}
		set
		{
			Columns.RowSize = value;
		}
	}

	public bool EOF
	{
		get
		{
			if (Cursor < Count)
			{
				return Count == 0;
			}
			return true;
		}
	}

	public bool First
	{
		get
		{
			if (Count > 0)
			{
				return Cursor == 0;
			}
			return false;
		}
	}

	public bool Last
	{
		get
		{
			if (Count > 0)
			{
				return Cursor == Count - 1;
			}
			return false;
		}
	}

	public bool IsEmpty
	{
		get
		{
			if (Count <= 0)
			{
				return true;
			}
			return false;
		}
	}

	[DebuggerBrowsable(DebuggerBrowsableState.Never)]
	private string DebuggerDisplay
	{
		get
		{
			StringBuilder stringBuilder = new StringBuilder();
			stringBuilder.Append((Name.Length == 0) ? "None" : Name);
			stringBuilder.Append(" count " + Count);
			if (_maxcount > 0)
			{
				stringBuilder.Append(" max " + _maxcount);
			}
			if (_pagesize > 0)
			{
				stringBuilder.Append(" page " + Page + " per " + _pagesize);
			}
			return stringBuilder.ToString();
		}
	}

	public static Record Split(string value)
	{
		return Split("name", value, ',');
	}

	public static Record Split(string name, string value)
	{
		return Split(name, value, ',');
	}

	public static Record Split(string name, string value, char spacer)
	{
		if (value == null || value.Length == 0)
		{
			return new Record();
		}
		string[] array = value.Split(new char[1] { spacer });
		Record record = new Record();
		record.Columns.Add(name);
		for (int i = 0; i < array.Length; i++)
		{
			record.AddRow();
			record[name].Set(array[i].Trim());
		}
		return record;
	}

	public static Record SplitInt(string value)
	{
		return SplitInt("name", value, ',');
	}

	public static Record SplitInt(string name, string value)
	{
		return SplitInt(name, value, ',');
	}

	public static Record SplitInt(string name, string value, char spacer)
	{
		if (value == null || value.Length == 0)
		{
			return new Record();
		}
		int[] array = ZYing.Split.Ints(value, spacer);
		Record record = new Record();
		record.Columns.AddInt(name);
		for (int i = 0; i < array.Length; i++)
		{
			record.AddRow();
			record[name].Set(array[i]);
		}
		return record;
	}

	public static Record ParseJson(string str)
	{
		return ParseJson(str, ',');
	}

	public static Record ParseJson(string str, char spacer)
	{
		return ParseJson(str, spacer, "key", "value");
	}

	public static Record ParseJson(string str, char spacer, string keyname, string vname)
	{
		Record record = new Record();
		try
		{
			str = str.Trim();
			if (str.StartsWith("{") && str.EndsWith("}"))
			{
				str = str.Substring(1, str.Length - 2);
			}
			record.Columns.Add(keyname);
			record.Columns.Add(vname);
			string[] array = str.Split(new char[1] { spacer });
			for (int i = 0; i < array.Length; i++)
			{
				string[] array2 = array[i].Split(new char[1] { ':' });
				if (array2.Length <= 1)
				{
					continue;
				}
				string text = array2[0].Trim();
				if (text.Length != 0)
				{
					if (text.Length > 1 && (text[0] == '"' || text[0] == '\''))
					{
						text = text.Substring(1, text.Length - 2);
					}
					string text2 = array2[1].Trim();
					if (text2.Length > 1 && (text2[0] == '"' || text2[0] == '\''))
					{
						text2 = text2.Substring(1, text2.Length - 2);
					}
					record.AddRow();
					record[keyname].Set(text);
					record[vname].Set(text2);
				}
			}
		}
		catch
		{
		}
		return record;
	}

	public static Record FromJson(string json)
	{
		Record record = new Record();
		if (string.IsNullOrEmpty(json))
		{
			return record;
		}
		JsonElement jsonElement = JsonElement.Create(json);
		if (jsonElement.ElementType == JsonElementType.Object)
		{
			ParseJson(record, (JsonObject)jsonElement);
		}
		else if (jsonElement.ElementType == JsonElementType.Array)
		{
			foreach (JsonObject item in (JsonArray)jsonElement)
			{
				record.AddRow();
				ParseJson(record, item);
			}
		}
		return record;
	}

	private static void ParseJson(Record re, JsonObject jObj)
	{
		foreach (string property in jObj.Properties)
		{
			JsonElement jsonElement = jObj[property];
			switch (jsonElement.ElementType)
			{
			case JsonElementType.String:
			{
				JsonString jsonString = (JsonString)jsonElement;
				re.Columns.AddString(property).Set(jsonString.Value);
				break;
			}
			case JsonElementType.Number:
			{
				JsonNumber jsonNumber = (JsonNumber)jsonElement;
				long value2;
				double value3;
				decimal value4;
				if (jsonNumber.TryGetInt(out var value))
				{
					re.Columns.AddInt(property).Set(value);
				}
				else if (jsonNumber.TryGetLong(out value2))
				{
					re.Columns.AddLong(property).Set(value2);
				}
				else if (jsonNumber.TryGetDouble(out value3))
				{
					re.Columns.AddDouble(property).Set(value3);
				}
				else if (jsonNumber.TryGetDecimal(out value4))
				{
					re.Columns.AddDecimal(property).Set(value4);
				}
				else
				{
					re.Columns.AddString(property).Set(jsonNumber.ToString());
				}
				break;
			}
			case JsonElementType.Boolean:
			{
				JsonBoolean jsonBoolean = (JsonBoolean)jsonElement;
				re.Columns.AddBool(property).Set(jsonBoolean.Value);
				break;
			}
			case JsonElementType.Array:
			{
				JsonArray jArr = (JsonArray)jsonElement;
				ParseJson(re, property, jArr);
				break;
			}
			}
		}
	}

	private static void ParseJson(Record re, string column, JsonArray jArr)
	{
		Array array = null;
		for (int i = 0; i < jArr.Count; i++)
		{
			JsonElement jsonElement = jArr[i];
			switch (jsonElement.ElementType)
			{
			case JsonElementType.String:
				if (i == 0)
				{
					array = new string[jArr.Count];
					re.Columns.AddStrings(column);
				}
				array.SetValue(jsonElement.ToString(), i);
				break;
			case JsonElementType.Number:
			{
				JsonNumber jsonNumber = (JsonNumber)jsonElement;
				long value2;
				decimal value3;
				if (jsonNumber.TryGetInt(out var value))
				{
					if (i == 0)
					{
						array = new int[jArr.Count];
						re.Columns.AddInts(column);
					}
					array.SetValue(value, i);
				}
				else if (jsonNumber.TryGetLong(out value2))
				{
					if (i == 0)
					{
						array = new long[jArr.Count];
						re.Columns.AddLongs(column);
					}
					array.SetValue(value2, i);
				}
				else if (jsonNumber.TryGetDecimal(out value3))
				{
					if (i == 0)
					{
						array = new decimal[jArr.Count];
						re.Columns.AddDecimals(column);
					}
					array.SetValue(value3, i);
				}
				break;
			}
			}
		}
		if (array != null)
		{
			re.Columns[column].Set(array);
		}
	}

	public static Record ToRows(NameValueCollection nv)
	{
		return ToRows(nv, "key", "value");
	}

	public static Record ToRows(NameValueCollection nv, string kname, string vname)
	{
		Record record = new Record();
		record.Columns.Add(kname);
		record.Columns.Add(vname);
		foreach (string key in nv.Keys)
		{
			record.AddRow();
			record[kname].Set(nv[key]);
		}
		return record;
	}

	public static Record ToCols(NameValueCollection nv)
	{
		Record record = new Record();
		foreach (string key in nv.Keys)
		{
			record.Columns.Add(key).Set(nv[key]);
		}
		return record;
	}

	public static Record From<T>(List<T> list) where T : new()
	{
		return From((IList<T>)list);
	}

	public static Record From<T>(IList<T> list) where T : new()
	{
		if (list.Count == 0)
		{
			return new Record();
		}
		if (list.Count == 1)
		{
			return From(list[0]);
		}
		string[] fields = XData<T>.Fields;
		ColumnAttribute[] columns = XData<T>.Columns;
		Record record = new Record(list.Count);
		Type[] array = null;
		for (int i = 0; i < fields.Length; i++)
		{
			ColumnAttribute columnAttribute = columns[i];
			if (columnAttribute != null)
			{
				TypeCode code = columnAttribute.DbType.ToCode(out var isArr);
				record.Columns.Add(columnAttribute.Name ?? fields[i], code, isArr);
				continue;
			}
			if (array == null)
			{
				array = XData<T>.GetTypes();
			}
			record.Columns.Add(fields[i], array[i]);
		}
		object[] array2 = new object[fields.Length];
		for (int j = 0; j < list.Count; j++)
		{
			record.AddRow();
			XData<T>.Output(list[j], array2);
			for (int k = 0; k < fields.Length; k++)
			{
				record.Columns[k, j].Set(array2[k]);
			}
		}
		return record;
	}

	public static Record From<T>(T t) where T : new()
	{
		Record record = new Record();
		record.Add(t);
		return record;
	}

	public Record Add<T>(T t) where T : new()
	{
		string[] fields = XData<T>.Fields;
		ColumnAttribute[] columns = XData<T>.Columns;
		object[] array = new object[fields.Length];
		XData<T>.Output(t, array);
		AddRow();
		Type[] array2 = null;
		for (int i = 0; i < fields.Length; i++)
		{
			ColumnAttribute columnAttribute = columns[i];
			if (columnAttribute != null)
			{
				TypeCode code = columnAttribute.DbType.ToCode(out var isArr);
				Columns.Add(columnAttribute.Name ?? fields[i], code, isArr).Set(array[i]);
				continue;
			}
			if (array2 == null)
			{
				array2 = XData<T>.GetTypes();
			}
			Columns.Add(fields[i], array2[i]).Set(array[i]);
		}
		return this;
	}

	public Record()
		: this(null, 20)
	{
	}

	public Record(int rows)
		: this(null, rows)
	{
	}

	public Record(string name)
		: this(name, 20)
	{
	}

	public Record(string name, int rows)
	{
		if (name != null)
		{
			Name = name;
		}
		if (rows <= 0)
		{
			rows = 20;
		}
		_cols = new RecordColumns(this, rows);
	}

	private Record(SerializationInfo info, StreamingContext ctxt)
	{
		string[] array = info.GetValue("ns", typeof(Array)) as string[];
		sbyte[] array2 = info.GetValue("ts", typeof(Array)) as sbyte[];
		Name = info.GetString("n");
		Count = info.GetInt32("k");
		Page = info.GetInt32("p");
		_maxcount = info.GetInt32("m");
		_pagesize = info.GetInt32("sz");
		_cols = new RecordColumns(this, info.GetInt32("rz"));
		for (int i = 0; i < array.Length; i++)
		{
			int num = array2[i];
			bool flag = num < 0;
			if (flag)
			{
				num = -num;
			}
			RecordColumn col = new RecordColumn(this, array[i], (TypeCode)num, flag)
			{
				Data = (Array)info.GetValue(i.ToString(), typeof(Array))
			};
			Columns.Add(col);
		}
	}

	public void GetObjectData(SerializationInfo info, StreamingContext context)
	{
		info.AddValue("n", Name);
		info.AddValue("k", Count);
		info.AddValue("p", Page);
		info.AddValue("m", _maxcount);
		info.AddValue("sz", _pagesize);
		info.AddValue("rz", RowSize);
		string[] array = new string[Columns.Count];
		sbyte[] array2 = new sbyte[Columns.Count];
		int num = 0;
		foreach (RecordColumn column in Columns)
		{
			info.AddValue(num.ToString(), column.Data);
			array[num] = column.Name;
			sbyte b = (sbyte)column.TypeCode;
			array2[num] = (column.IsArray ? ((sbyte)(-b)) : b);
			num++;
		}
		info.AddValue("ns", array);
		info.AddValue("ts", array2);
	}

	public Record Row2Col(string a, string b)
	{
		Record record = new Record();
		RecordColumn recordColumn = this[a];
		RecordColumn recordColumn2 = this[b];
		for (int i = 0; i < Count; i++)
		{
			record.Columns.Add(recordColumn.Name + "_" + recordColumn[i].ToString(), recordColumn.TypeCode, recordColumn.IsArray).Set(recordColumn2[i].ToObject());
		}
		_cols = record.Columns;
		Count = record.Count;
		return this;
	}

	public Record Combine(Record re)
	{
		foreach (RecordColumn column in Columns)
		{
			RecordColumn recordColumn = re[column.Name];
			if (recordColumn != null)
			{
				for (int i = 0; i < Count && i < re.Count; i++)
				{
					column[i].Add(recordColumn[i].ToLong());
				}
			}
		}
		return this;
	}

	public Record Union(Record re)
	{
		int count = re.Count;
		if (count == 0)
		{
			return this;
		}
		int count2 = re.Columns.Count;
		RecordColumn[] array = new RecordColumn[count2];
		for (int i = 0; i < count2; i++)
		{
			RecordColumn recordColumn = re.Columns[i];
			array[i] = Columns.Add(recordColumn.Name, recordColumn.TypeCode, recordColumn.IsArray);
		}
		RowSize += Count + count;
		for (int j = 0; j < count2; j++)
		{
			RecordColumn col = re.Columns[j];
			array[j].Union(col);
		}
		Count += count;
		return this;
	}

	public string Join(params string[] ns)
	{
		StringBuilder stringBuilder = new StringBuilder();
		foreach (string name in ns)
		{
			string text = Get(name).Trim();
			if (text.Length != 0)
			{
				if (stringBuilder.Length > 0)
				{
					stringBuilder.Append(' ');
				}
				stringBuilder.Append(text);
			}
		}
		return stringBuilder.ToString();
	}

	public Record Join(Record re)
	{
		foreach (RecordColumn column in re.Columns)
		{
			Columns.Add(column.Name, column.TypeCode, column.IsArray).Set(column[Cursor]);
		}
		return this;
	}

	public Record Join(Record re, string fldname)
	{
		return Join(re, fldname, fldname);
	}

	public Record Join(Record re, string selfName, string otherName)
	{
		int count = re.Count;
		if (count == 0 || !Contains(selfName) || !re.Contains(otherName))
		{
			return this;
		}
		MoveFirst();
		foreach (RecordColumn column in re.Columns)
		{
			Columns.Add(column.Name, column.TypeCode, column.IsArray);
		}
		RecordColumn recordColumn = this[selfName];
		RecordColumn recordColumn2 = re[otherName];
		for (int i = 0; i < Count; i++)
		{
			for (int j = 0; j < count; j++)
			{
				if (!recordColumn[i].Equals(recordColumn2[j]))
				{
					continue;
				}
				foreach (RecordColumn column2 in re.Columns)
				{
					if (column2.Name != selfName && !column2[j].IsEmpty)
					{
						this[column2.Name, i].Set(column2[j]);
					}
				}
				break;
			}
		}
		return this;
	}

	public Record JoinEmpty(Record re, string fieldname)
	{
		return JoinEmpty(re, fieldname, fieldname);
	}

	public Record JoinEmpty(Record re, string self, string other)
	{
		if (re.Count == 0 || !Contains(self) || !re.Contains(other))
		{
			return this;
		}
		MoveFirst();
		foreach (RecordColumn column in re.Columns)
		{
			Columns.Add(column.Name, column.TypeCode, column.IsArray);
		}
		while (Read())
		{
			for (int i = 0; i < re.Count; i++)
			{
				if (!this[self].Equals(re[other, i]))
				{
					continue;
				}
				foreach (RecordColumn column2 in re.Columns)
				{
					string text = this[column2.Name].ToString();
					if (text.Length == 0 || text == "0" || text == "0.00")
					{
						this[column2.Name].Set(column2[i]);
					}
				}
				break;
			}
		}
		return this;
	}

	public Record InnerJoin(Record re, string fieldname)
	{
		return InnerJoin(re, fieldname, fieldname);
	}

	public Record InnerJoin(Record re, string selfName, string otherName)
	{
		if (Count == 0 || !Contains(selfName) || !re.Contains(otherName))
		{
			return this;
		}
		int count = re.Count;
		if (count == 0)
		{
			Clear();
		}
		else
		{
			MoveFirst();
			foreach (RecordColumn column in re.Columns)
			{
				Columns.Add(column.Name, column.TypeCode, column.IsArray);
			}
			RecordColumn recordColumn = this[selfName];
			RecordColumn recordColumn2 = re[otherName];
			for (int num = Count - 1; num >= 0; num--)
			{
				bool flag = false;
				for (int i = 0; i < count; i++)
				{
					if (!recordColumn[num].Equals(recordColumn2[i]))
					{
						continue;
					}
					foreach (RecordColumn column2 in re.Columns)
					{
						if (column2.Name != selfName && !column2[i].IsEmpty)
						{
							this[column2.Name, num].Set(column2[i]);
						}
						flag = true;
					}
					break;
				}
				if (!flag)
				{
					Delete(num);
				}
			}
		}
		return this;
	}

	public int Sum(string name)
	{
		if (!Contains(name))
		{
			return 0;
		}
		int num = 0;
		for (int i = 0; i < Count; i++)
		{
			num += this[name, i].ToInt();
		}
		return num;
	}

	public long SumLong(string name)
	{
		if (!Contains(name))
		{
			return 0L;
		}
		long num = 0L;
		for (int i = 0; i < Count; i++)
		{
			num += this[name, i].ToLong();
		}
		return num;
	}

	public double SumDouble(string name)
	{
		if (!Contains(name))
		{
			return 0.0;
		}
		double num = 0.0;
		for (int i = 0; i < Count; i++)
		{
			num += this[name, i].ToDouble();
		}
		return num;
	}

	public decimal SumDecimal(string name)
	{
		if (!Contains(name))
		{
			return 0m;
		}
		decimal result = 0m;
		for (int i = 0; i < Count; i++)
		{
			result += this[name, i].ToDecimal();
		}
		return result;
	}

	public int Min(string name)
	{
		if (!Contains(name))
		{
			return 0;
		}
		int num = int.MaxValue;
		for (int i = 0; i < Count; i++)
		{
			int num2 = this[name, i].ToInt();
			if (num2 < num)
			{
				num = num2;
			}
		}
		if (num != int.MaxValue)
		{
			return num;
		}
		return 0;
	}

	public int Max(string name)
	{
		if (!Contains(name))
		{
			return 0;
		}
		int num = 0;
		for (int i = 0; i < Count; i++)
		{
			int num2 = this[name, i].ToInt();
			if (num2 > num)
			{
				num = num2;
			}
		}
		return num;
	}

	public long MinLong(string name)
	{
		if (!Contains(name))
		{
			return 0L;
		}
		long num = long.MaxValue;
		for (int i = 0; i < Count; i++)
		{
			long num2 = this[name, i].ToLong();
			if (num2 < num)
			{
				num = num2;
			}
		}
		if (num != long.MaxValue)
		{
			return num;
		}
		return 0L;
	}

	public long MaxLong(string name)
	{
		if (!Contains(name))
		{
			return 0L;
		}
		long num = 0L;
		for (int i = 0; i < Count; i++)
		{
			long num2 = this[name, i].ToLong();
			if (num2 > num)
			{
				num = num2;
			}
		}
		return num;
	}

	public decimal MinDecimal(string name)
	{
		if (!Contains(name))
		{
			return 0m;
		}
		decimal num = decimal.MaxValue;
		for (int i = 0; i < Count; i++)
		{
			decimal num2 = this[name, i].ToDecimal();
			if (num2 < num)
			{
				num = num2;
			}
		}
		if (!(num == decimal.MaxValue))
		{
			return num;
		}
		return 0m;
	}

	public decimal MaxDecimal(string name)
	{
		if (!Contains(name))
		{
			return 0m;
		}
		decimal num = 0m;
		for (int i = 0; i < Count; i++)
		{
			decimal num2 = this[name, i].ToDecimal();
			if (num2 > num)
			{
				num = num2;
			}
		}
		return num;
	}

	public float MaxFloat(params string[] ns)
	{
		float num = 0f;
		for (int i = 0; i < Count; i++)
		{
			float num2 = 0f;
			foreach (string name in ns)
			{
				num2 += this[name, i].ToFloat();
			}
			if (num2 > num)
			{
				num = num2;
			}
		}
		return num;
	}

	public float MinFloat(params string[] ns)
	{
		float num = float.MaxValue;
		for (int i = 0; i < Count; i++)
		{
			float num2 = 0f;
			foreach (string name in ns)
			{
				num2 += this[name, i].ToFloat();
			}
			if (num2 < num)
			{
				num = num2;
			}
		}
		if (num != float.MaxValue)
		{
			return num;
		}
		return 0f;
	}

	public bool Contains(string name)
	{
		return Columns.Contains(name);
	}

	public string GetFirst(string name)
	{
		while (Read())
		{
			string text = Get(name);
			if (text.Length > 0)
			{
				MoveFirst();
				return text;
			}
		}
		return null;
	}

	public string Get(string name)
	{
		if (!Contains(name))
		{
			return "";
		}
		return this[name].ToString().Trim();
	}

	public string[] Gets(string name)
	{
		if (!Contains(name))
		{
			return new string[0];
		}
		return this[name].ToStrings();
	}

	public Guid GetGuid(string name)
	{
		if (!Contains(name))
		{
			return Guid.Empty;
		}
		return this[name].ToGuid();
	}

	public long GetLong(string name)
	{
		if (!Contains(name))
		{
			return 0L;
		}
		return this[name].ToLong();
	}

	public long[] GetLongs(string name)
	{
		if (!Contains(name))
		{
			return new long[0];
		}
		return this[name].ToLongs();
	}

	public byte[] GetBytes(string name)
	{
		if (!Contains(name))
		{
			return new byte[0];
		}
		return this[name].ToBytes();
	}

	public int GetInt(string name)
	{
		if (!Contains(name))
		{
			return 0;
		}
		return this[name].ToInt();
	}

	public int[] GetInts(string name)
	{
		if (!Contains(name))
		{
			return new int[0];
		}
		return this[name].ToInts();
	}

	public decimal GetMoney(string name)
	{
		if (!Contains(name))
		{
			return 0m;
		}
		return this[name].ToDecimal();
	}

	public decimal[] GetMoneys(string name)
	{
		if (!Contains(name))
		{
			return new decimal[0];
		}
		return this[name].ToDecimals();
	}

	public float GetFloat(string name)
	{
		if (!Contains(name))
		{
			return 0f;
		}
		return this[name].ToFloat();
	}

	public double GetDouble(string name)
	{
		if (!Contains(name))
		{
			return 0.0;
		}
		return this[name].ToDouble();
	}

	public DateTime GetDate(string name)
	{
		if (!Contains(name))
		{
			return DateTime.MinValue;
		}
		return this[name].ToDate();
	}

	public string GetDateText(string name, string format = "yyyy-MM-dd HH:mm:ss")
	{
		DateTime date = GetDate(name);
		if (!(date == DateTime.MinValue))
		{
			return date.ToString(format);
		}
		return "";
	}

	public bool GetBool(string name)
	{
		if (!Contains(name))
		{
			return false;
		}
		return this[name].ToBool();
	}

	public void ReName(string old, string name)
	{
		Columns.ReName(old, name);
	}

	public Record Prefix(string pre)
	{
		return Prefix("", pre);
	}

	public Record Prefix(string old, string pre)
	{
		Columns.Prefix(old, pre);
		return this;
	}

	public Record Abbr()
	{
		foreach (RecordColumn column in Columns)
		{
			int num = column.Name.IndexOf('_');
			if (num > 0)
			{
				Columns.ReName(column.Name, column.Name.Substring(num + 1));
			}
		}
		return this;
	}

	public void ChangeType(string name, TypeCode code, bool isarr = false)
	{
		Columns[name]?.ChangeType(code, isarr);
	}

	public bool Read()
	{
		if (_reading == OrderBy.ASC)
		{
			if (Cursor >= Count - 1)
			{
				_reading = OrderBy.None;
				Cursor = 0;
				return false;
			}
			Cursor++;
		}
		else
		{
			if (_reading != OrderBy.None)
			{
				throw new Exception("Record.Read()期间，_reading状态为倒序！");
			}
			if (Count == 0)
			{
				return false;
			}
			Cursor = 0;
			_reading = OrderBy.ASC;
		}
		return true;
	}

	public bool Read(string name, int value)
	{
		while (Read())
		{
			if (this[name].ToInt() == value)
			{
				return true;
			}
		}
		return false;
	}

	public bool Read(string name, long value)
	{
		while (Read())
		{
			if (this[name].ToLong() == value)
			{
				return true;
			}
		}
		return false;
	}

	public bool Read(string name, string value)
	{
		if (value == null)
		{
			value = "";
		}
		while (Read())
		{
			if (string.Compare(this[name].ToString(), value, ignoreCase: true) == 0)
			{
				return true;
			}
		}
		return false;
	}

	public bool ReadBack()
	{
		if (_reading == OrderBy.DESC)
		{
			if (Cursor <= 0)
			{
				_reading = OrderBy.None;
				Cursor = 0;
				return false;
			}
			Cursor--;
		}
		else
		{
			if (_reading != OrderBy.None)
			{
				throw new Exception("Record.ReadBack()期间，_reading状态为正序！");
			}
			if (Count == 0)
			{
				return false;
			}
			Cursor = Count - 1;
			_reading = OrderBy.DESC;
		}
		return true;
	}

	public void MoveFirst()
	{
		_reading = OrderBy.None;
		Cursor = 0;
	}

	public void MoveLast()
	{
		_reading = OrderBy.None;
		Cursor = Count - 1;
	}

	public void MoveNext()
	{
		if (!EOF)
		{
			Cursor++;
		}
	}

	public void Clear()
	{
		Count = 0;
		_cols = new RecordColumns(this, _cols.RowSize);
	}

	public void Reset()
	{
		foreach (RecordColumn column in Columns)
		{
			column.Reset(Cursor);
		}
	}

	public void AddRow()
	{
		Cursor = Count;
		RowSize = ++Count;
	}

	public void AddRow(Record re)
	{
		AddRow();
		foreach (RecordColumn column in re.Columns)
		{
			Columns.Add(column.Name, column.TypeCode, column.IsArray).Set(column[re.Cursor]);
		}
	}

	public void InsertRow()
	{
		InsertRow(0);
	}

	public void InsertRow(int index)
	{
		if (index < Count)
		{
			Columns.InsertRow(index);
			Cursor = index;
			int count = Count + 1;
			Count = count;
		}
		else
		{
			AddRow();
		}
	}

	public bool Delete()
	{
		return Delete(Cursor);
	}

	public bool Delete(int row)
	{
		if (row < 0 || row >= Count)
		{
			return false;
		}
		foreach (RecordColumn column in Columns)
		{
			Array data = column.Data;
			for (int i = row; i < Count - 1; i++)
			{
				data.SetValue(data.GetValue(i + 1), i);
			}
		}
		Count--;
		return true;
	}

	public void CreateRank()
	{
		CreateRank("rank");
	}

	public void CreateRank(string colname)
	{
		Columns.AddInt(colname);
		for (int i = 0; i < Count; i++)
		{
			this[colname, i].Set(i + 1);
		}
	}

	public void SetParagraph(string name)
	{
		RecordColumn recordColumn = Columns.Get(name);
		if (recordColumn != null)
		{
			for (int i = 0; i < Count; i++)
			{
				recordColumn[i].Paragraph();
			}
		}
	}

	public void SetArticle(string name)
	{
		RecordColumn recordColumn = Columns.Get(name);
		if (recordColumn != null)
		{
			for (int i = 0; i < Count; i++)
			{
				recordColumn[i].Article();
			}
		}
	}

	public void Escape(string name)
	{
		RecordColumn recordColumn = Columns.Get(name);
		if (recordColumn != null)
		{
			for (int i = 0; i < Count; i++)
			{
				recordColumn[i].Escape();
			}
		}
	}

	public RecordRow FindIn(string field, int value)
	{
		RecordColumn recordColumn = this[field];
		if (recordColumn == null)
		{
			return null;
		}
		for (int i = 0; i < Count; i++)
		{
			int[] array = ZYing.Split.Ints(recordColumn[i].ToString());
			for (int j = 0; j < array.Length; j++)
			{
				if (array[j] == value)
				{
					return new RecordRow(this, i);
				}
			}
		}
		return null;
	}

	public RecordRow Find(string name, bool value)
	{
		int num = FindIndex(name, value);
		if (num != -1)
		{
			return new RecordRow(this, num);
		}
		return null;
	}

	public RecordRow Find(string name, int value)
	{
		int num = FindIndex(name, value);
		if (num != -1)
		{
			return new RecordRow(this, num);
		}
		return null;
	}

	public RecordRow Find(string name, byte value)
	{
		int num = FindIndex(name, value);
		if (num != -1)
		{
			return new RecordRow(this, num);
		}
		return null;
	}

	public RecordRow Find(string name, long value)
	{
		int num = FindIndex(name, value);
		if (num != -1)
		{
			return new RecordRow(this, num);
		}
		return null;
	}

	public RecordRow Find(string name, float value)
	{
		int num = FindIndex(name, value);
		if (num != -1)
		{
			return new RecordRow(this, num);
		}
		return null;
	}

	public RecordRow Find(string name, double value)
	{
		int num = FindIndex(name, value);
		if (num != -1)
		{
			return new RecordRow(this, num);
		}
		return null;
	}

	public RecordRow Find(string name, DateTime value)
	{
		int num = FindIndex(name, value);
		if (num != -1)
		{
			return new RecordRow(this, num);
		}
		return null;
	}

	public RecordRow Find(string name, decimal value)
	{
		int num = FindIndex(name, value);
		if (num != -1)
		{
			return new RecordRow(this, num);
		}
		return null;
	}

	public RecordRow Find(string name, string value)
	{
		int num = FindIndex(name, value);
		if (num != -1)
		{
			return new RecordRow(this, num);
		}
		return null;
	}

	public RecordRow Find(string name, object value)
	{
		int num = FindIndex(name, value);
		if (num != -1)
		{
			return new RecordRow(this, num);
		}
		return null;
	}

	public RecordRows FindList(string name, int value)
	{
		RecordRows recordRows = new RecordRows();
		RecordColumn recordColumn = this[name];
		if (recordColumn == null)
		{
			return recordRows;
		}
		for (int i = 0; i < Count; i++)
		{
			if (recordColumn[i].Equals(value))
			{
				RecordRow item = new RecordRow(this, i);
				recordRows.Add(item);
			}
		}
		return recordRows;
	}

	public int FindIndex(string name, bool value)
	{
		RecordColumn recordColumn = this[name];
		if (recordColumn == null)
		{
			return -1;
		}
		for (int i = 0; i < Count; i++)
		{
			if (recordColumn[i].Equals(value))
			{
				return i;
			}
		}
		return -1;
	}

	public int FindIndex(string name, int value)
	{
		RecordColumn recordColumn = this[name];
		if (recordColumn == null)
		{
			return -1;
		}
		for (int i = 0; i < Count; i++)
		{
			if (recordColumn[i].Equals(value))
			{
				return i;
			}
		}
		return -1;
	}

	public int FindIndex(string name, byte value)
	{
		RecordColumn recordColumn = this[name];
		if (recordColumn == null)
		{
			return -1;
		}
		for (int i = 0; i < Count; i++)
		{
			if (recordColumn[i].Equals(value))
			{
				return i;
			}
		}
		return -1;
	}

	public int FindIndex(string name, long value)
	{
		RecordColumn recordColumn = this[name];
		if (recordColumn == null)
		{
			return -1;
		}
		for (int i = 0; i < Count; i++)
		{
			if (recordColumn[i].Equals(value))
			{
				return i;
			}
		}
		return -1;
	}

	public int FindIndex(string name, float value)
	{
		RecordColumn recordColumn = this[name];
		if (recordColumn == null)
		{
			return -1;
		}
		for (int i = 0; i < Count; i++)
		{
			if (recordColumn[i].Equals(value))
			{
				return i;
			}
		}
		return -1;
	}

	public int FindIndex(string name, double value)
	{
		RecordColumn recordColumn = this[name];
		if (recordColumn == null)
		{
			return -1;
		}
		for (int i = 0; i < Count; i++)
		{
			if (recordColumn[i].Equals(value))
			{
				return i;
			}
		}
		return -1;
	}

	public int FindIndex(string name, DateTime value)
	{
		RecordColumn recordColumn = this[name];
		if (recordColumn == null)
		{
			return -1;
		}
		for (int i = 0; i < Count; i++)
		{
			if (recordColumn[i].Equals(value))
			{
				return i;
			}
		}
		return -1;
	}

	public int FindIndex(string name, decimal value)
	{
		RecordColumn recordColumn = this[name];
		if (recordColumn == null)
		{
			return -1;
		}
		for (int i = 0; i < Count; i++)
		{
			if (recordColumn[i].Equals(value))
			{
				return i;
			}
		}
		return -1;
	}

	public int FindIndex(string name, string value)
	{
		RecordColumn recordColumn = this[name];
		if (recordColumn == null)
		{
			return -1;
		}
		for (int i = 0; i < Count; i++)
		{
			if (recordColumn[i].Equals(value))
			{
				return i;
			}
		}
		return -1;
	}

	public int FindIndex(string name, object value)
	{
		RecordColumn recordColumn = this[name];
		if (recordColumn == null)
		{
			return -1;
		}
		for (int i = 0; i < Count; i++)
		{
			if (recordColumn[i].Equals(value))
			{
				return i;
			}
		}
		return -1;
	}

	public void Fill(Record dest, params string[] fldname)
	{
		if (fldname.Length == 0)
		{
			foreach (RecordColumn column in Columns)
			{
				if (!dest.Contains(column.Name))
				{
					RecordColumn recordColumn = dest.Columns.Add(column.Name, column.TypeCode, column.IsArray);
					Array.Copy(column.Data, 0, recordColumn.Data, 0, recordColumn.Data.GetLength(0));
				}
			}
			return;
		}
		foreach (string name in fldname)
		{
			if (!dest.Contains(name) && Contains(name))
			{
				RecordColumn recordColumn2 = this[name];
				RecordColumn recordColumn3 = dest.Columns.Add(recordColumn2.Name, recordColumn2.TypeCode, recordColumn2.IsArray);
				Array.Copy(recordColumn2.Data, 0, recordColumn3.Data, 0, recordColumn3.Data.GetLength(0));
			}
		}
	}

	public void Fill(string field, object value)
	{
		Columns.Add(field).Fill(value);
	}

	public void Fill(string field, Transfer.Object tran)
	{
		Columns.Add(field).Replace(tran);
	}

	public void Fill(string field, Transfer.Line tran)
	{
		Columns.Add(field).Replace(tran);
	}

	public void Fill(string newfld, string name, Type etype)
	{
		if (Contains(name))
		{
			Columns.Add(newfld);
			for (int i = 0; i < Count; i++)
			{
				string name2 = Valid.GetName(this[name, i].ToInt(), etype);
				this[newfld, i].Set(name2);
			}
		}
	}

	public void Fill(string newfld, string name, Transfer.Object tran)
	{
		if (Contains(name))
		{
			Columns.Add(newfld);
			for (int i = 0; i < Count; i++)
			{
				this[newfld, i].Set(tran(this[name, i].ToObject()));
			}
		}
	}

	public void Fill(string newfld, string name, Transfer.String tran)
	{
		if (Contains(name))
		{
			Columns.Add(newfld);
			for (int i = 0; i < Count; i++)
			{
				this[newfld, i].Set(tran(this[name, i].ToString()));
			}
		}
	}

	public void Fill(string newfld, string name, Transfer.StringInt tran)
	{
		if (Contains(name))
		{
			Columns.Add(newfld);
			for (int i = 0; i < Count; i++)
			{
				this[newfld, i].Set(tran(this[name, i].ToInt()));
			}
		}
	}

	public void Fill(string newfld, string name, Transfer.Long tran)
	{
		if (Contains(name))
		{
			Columns.AddLong(newfld);
			for (int i = 0; i < Count; i++)
			{
				this[newfld, i].Set(tran(this[name, i].ToLong()));
			}
		}
	}

	public void Fill(string newfld, string name, Transfer.Double tran)
	{
		if (Contains(name))
		{
			Columns.AddDouble(newfld);
			for (int i = 0; i < Count; i++)
			{
				this[newfld, i].Set(tran(this[name, i].ToDouble()));
			}
		}
	}

	public void Fill(string newfld, string name, Transfer.DateTime tran)
	{
		if (Contains(name))
		{
			Columns.AddDate(newfld);
			for (int i = 0; i < Count; i++)
			{
				this[newfld, i].Set(tran(this[name, i].ToDate()));
			}
		}
	}

	public void Replace(string name, string old, string value)
	{
		if (Contains(name))
		{
			this[name].Replace(old, value);
		}
	}

	public void Replace(string name, Transfer.DateTime tran)
	{
		Columns.Get(name)?.Replace(tran);
	}

	public void Replace(string name, Transfer.Double tran)
	{
		Columns.Get(name)?.Replace(tran);
	}

	public void Replace(string name, Transfer.Long tran)
	{
		Columns.Get(name)?.Replace(tran);
	}

	public void Replace(string name, Transfer.String tran)
	{
		Columns.Get(name)?.Replace(tran);
	}

	public void Replace(string name, Transfer.Object tran)
	{
		Columns.Get(name)?.Replace(tran);
	}

	public void Replace(string name, Transfer.Line tran)
	{
		Columns.Get(name)?.Replace(tran);
	}

	public Record Distinct(params string[] cols)
	{
		Record record = new Record();
		foreach (RecordColumn column in Columns)
		{
			record.Columns.Add(column.Name, column.TypeCode, column.IsArray);
		}
		Dictionary<string, bool> dictionary = new Dictionary<string, bool>(StringComparer.OrdinalIgnoreCase);
		while (Read())
		{
			StringBuilder stringBuilder = new StringBuilder();
			for (int i = 0; i < cols.Length; i++)
			{
				if (stringBuilder.Length > 0)
				{
					stringBuilder.Append("|");
				}
				stringBuilder.Append(cols[i] + "_" + this[cols[i]].ToString());
			}
			string key = stringBuilder.ToString();
			if (dictionary.ContainsKey(key))
			{
				continue;
			}
			dictionary.Add(key, value: true);
			record.AddRow();
			foreach (RecordColumn column2 in Columns)
			{
				record[column2.Name].Set(column2[Cursor]);
			}
		}
		return record;
	}

	public int[] DistinctInt(string col)
	{
		Dictionary<int, bool> dictionary = new Dictionary<int, bool>();
		while (Read())
		{
			int key = this[col].ToInt();
			dictionary[key] = true;
		}
		int[] array = new int[dictionary.Count];
		dictionary.Keys.CopyTo(array, 0);
		return array;
	}

	public Dictionary<string, List<long>> GroupStringLongs(string colname, string keyname)
	{
		Dictionary<string, List<long>> dictionary = new Dictionary<string, List<long>>(StringComparer.OrdinalIgnoreCase);
		if (!Contains(colname) || !Contains(keyname))
		{
			return dictionary;
		}
		while (Read())
		{
			string text = this[colname].ToString().Trim();
			if (text.Length != 0)
			{
				if (!dictionary.TryGetValue(text, out var value))
				{
					value = (dictionary[text] = new List<long>());
				}
				value.Add(this[keyname].ToLong());
			}
		}
		return dictionary;
	}

	public Dictionary<string, Record> Group(params string[] cols)
	{
		Dictionary<string, Record> dictionary = new Dictionary<string, Record>(StringComparer.OrdinalIgnoreCase);
		while (Read())
		{
			string text = "";
			foreach (string name in cols)
			{
				string text2 = this[name].ToString();
				text = text + ((text.Length > 0) ? "-" : "") + text2;
			}
			if (!dictionary.TryGetValue(text, out var value))
			{
				value = new Record();
				foreach (RecordColumn column in Columns)
				{
					value.Columns.Add(column.Name, column.TypeCode, column.IsArray);
				}
				dictionary[text] = value;
			}
			value.AddRow();
			foreach (RecordColumn column2 in Columns)
			{
				value[column2.Name].Set(column2[Cursor]);
			}
		}
		return dictionary;
	}

	public Dictionary<int, Record> GroupInt(string colname)
	{
		Dictionary<int, Record> dictionary = new Dictionary<int, Record>();
		while (Read())
		{
			int key = GetInt(colname);
			if (!dictionary.TryGetValue(key, out var value))
			{
				value = new Record();
				foreach (RecordColumn column in Columns)
				{
					value.Columns.Add(column.Name, column.TypeCode, column.IsArray);
				}
				dictionary[key] = value;
			}
			value.AddRow();
			foreach (RecordColumn column2 in Columns)
			{
				value[column2.Name].Set(column2[Cursor]);
			}
		}
		return dictionary;
	}

	public Dictionary<long, Record> GroupLong(string colname)
	{
		Dictionary<long, Record> dictionary = new Dictionary<long, Record>();
		while (Read())
		{
			long key = GetLong(colname);
			if (!dictionary.TryGetValue(key, out var value))
			{
				value = new Record();
				foreach (RecordColumn column in Columns)
				{
					value.Columns.Add(column.Name, column.TypeCode, column.IsArray);
				}
				dictionary[key] = value;
			}
			value.AddRow();
			foreach (RecordColumn column2 in Columns)
			{
				value[column2.Name].Set(column2[Cursor]);
			}
		}
		return dictionary;
	}

	public Dictionary<TKey, TValue> Group<TKey, TValue>(string colname, string vname)
	{
		Dictionary<TKey, TValue> dictionary = new Dictionary<TKey, TValue>();
		RecordColumn recordColumn = this[colname];
		if (recordColumn == null)
		{
			return dictionary;
		}
		RecordColumn recordColumn2 = this[vname];
		if (recordColumn2 == null)
		{
			return dictionary;
		}
		for (int i = 0; i < Count; i++)
		{
			TKey key = recordColumn[i].To<TKey>();
			dictionary[key] = recordColumn2.To<TValue>();
		}
		return dictionary;
	}

	public Dictionary<int, List<int>> GroupIntInts(string keyname, string vname)
	{
		Dictionary<int, List<int>> dictionary = new Dictionary<int, List<int>>();
		while (Read())
		{
			int num = this[keyname].ToInt();
			if (num > 0)
			{
				if (!dictionary.TryGetValue(num, out var value))
				{
					value = (dictionary[num] = new List<int>());
				}
				value.Add(GetInt(vname));
			}
		}
		return dictionary;
	}

	public Dictionary<int, List<long>> GroupIntLongs(string keyname, string vname)
	{
		Dictionary<int, List<long>> dictionary = new Dictionary<int, List<long>>();
		while (Read())
		{
			int num = this[keyname].ToInt();
			if (num > 0)
			{
				if (!dictionary.TryGetValue(num, out var value))
				{
					value = (dictionary[num] = new List<long>());
				}
				value.Add(GetLong(vname));
			}
		}
		return dictionary;
	}

	public Dictionary<string, int> GroupSum(string colname, string vname)
	{
		Dictionary<string, int> dictionary = new Dictionary<string, int>(StringComparer.OrdinalIgnoreCase);
		RecordColumn recordColumn = this[colname];
		if (recordColumn == null)
		{
			return dictionary;
		}
		RecordColumn recordColumn2 = this[vname];
		if (recordColumn2 == null)
		{
			return dictionary;
		}
		for (int i = 0; i < Count; i++)
		{
			string key = recordColumn[i].ToString();
			int num = recordColumn2[i].ToInt();
			dictionary[key] = (dictionary.TryGetValue(key, out var value) ? (value + num) : num);
		}
		return dictionary;
	}

	public Dictionary<string, int> GroupCount(string colname)
	{
		Dictionary<string, int> dictionary = new Dictionary<string, int>(StringComparer.OrdinalIgnoreCase);
		RecordColumn recordColumn = this[colname];
		if (recordColumn == null)
		{
			return dictionary;
		}
		for (int i = 0; i < Count; i++)
		{
			string key = recordColumn[i].ToString();
			dictionary[key] = ((!dictionary.TryGetValue(key, out var value)) ? 1 : (value + 1));
		}
		return dictionary;
	}

	public Record Filter(string name, int value)
	{
		Record record = new Record();
		foreach (RecordColumn column in Columns)
		{
			record.Columns.Add(column.Name, column.TypeCode, column.IsArray);
		}
		while (Read(name, value))
		{
			record.AddRow();
			foreach (RecordColumn column2 in Columns)
			{
				record[column2.Name].Set(column2[Cursor]);
			}
		}
		return record;
	}

	public Record Filter(string name, long value)
	{
		Record record = new Record();
		foreach (RecordColumn column in Columns)
		{
			record.Columns.Add(column.Name, column.TypeCode, column.IsArray);
		}
		while (Read(name, value))
		{
			record.AddRow();
			foreach (RecordColumn column2 in Columns)
			{
				record[column2.Name].Set(column2[Cursor]);
			}
		}
		return record;
	}

	public Record Extract(string field, int value)
	{
		Record record = new Record(Name);
		RecordColumn recordColumn = this[field];
		if (recordColumn == null)
		{
			return record;
		}
		for (int i = 0; i < Count; i++)
		{
			if (!recordColumn[i].Equals(value))
			{
				continue;
			}
			foreach (RecordColumn column in Columns)
			{
				record.Columns.Add(column.Name, column.TypeCode, column.IsArray).Set(column[i]);
			}
		}
		return record;
	}

	public string Shrink()
	{
		if (!Contains("name") || !Contains("value"))
		{
			throw new Exception("Record does not Contains name and value!");
		}
		StringBuilder stringBuilder = new StringBuilder();
		while (Read())
		{
			if (stringBuilder.Length > 0)
			{
				stringBuilder.Append(",");
			}
			stringBuilder.Append("\"" + Xml.Encode(this["name"].ToString()) + "\"");
			stringBuilder.Append(":");
			stringBuilder.Append("\"" + Xml.Encode(this["value"].ToString()) + "\"");
		}
		return stringBuilder.ToString();
	}

	public void Regex(string name, string RegStr, string value)
	{
		this[name].Reg(RegStr, value);
	}

	public void Regex(string name, string RegStr, MatchEvaluator Func)
	{
		this[name].Reg(RegStr, Func);
	}

	public void HighLight(string name)
	{
		HighLight(name, null);
	}

	public void HighLight(string name, string s)
	{
		if (Contains(name))
		{
			ChangeType(name, TypeCode.String);
			for (int i = 0; i < Count; i++)
			{
				this[name, i].HighLight(s);
			}
		}
	}

	public void Fill<T>(T t) where T : new()
	{
		ExecHeader<T>(out var subs);
		ExecValue(t, Cursor, subs);
	}

	public T To<T>() where T : new()
	{
		if (Count == 0)
		{
			return new T();
		}
		ExecHeader<T>(out var subs);
		return ExecValue<T>(default, Cursor, subs);
	}

	public object ToObject(Type type)
	{
		return typeof(Record).GetMethod("To").MakeGenericMethod(type).Invoke(this, null);
	}

	public List<T> ToList<T>() where T : new()
	{
		List<T> list = new List<T>();
		if (Count == 0)
		{
			return list;
		}
		ExecHeader<T>(out var subs);
		object[] tmp = new object[_cols.Count];
		object[] arr = new object[subs.Length];
		for (int i = 0; i < Count; i++)
		{
			T item = ExecValue<T>(default, i, subs, tmp, arr);
			list.Add(item);
		}
		return list;
	}

	public object ToListObject(Type type)
	{
		return typeof(Record).GetMethod("ToList").MakeGenericMethod(type).Invoke(this, null);
	}

	private void ExecHeader<T>(out int[] subs) where T : new()
	{
		string[] fields = XData<T>.Fields;
		ColumnAttribute[] columns = XData<T>.Columns;
		subs = new int[fields.Length];
		for (int i = 0; i < subs.Length; i++)
		{
			subs[i] = -1;
		}
		for (int j = 0; j < _cols.Count; j++)
		{
			string name = _cols[j].Name;
			int num = -1;
			for (int k = 0; k < columns.Length; k++)
			{
				ColumnAttribute columnAttribute = columns[k];
				if ((columnAttribute == null || !columnAttribute.Ignore) && ((columnAttribute == null || columnAttribute.Name == null) ? fields[k] : columnAttribute.Name) == name)
				{
					num = k;
					break;
				}
			}
			if (num >= 0)
			{
				subs[num] = j;
			}
		}
	}

	private T ExecValue<T>(T t, int line, int[] subs, object[] tmp = null, object[] arr = null) where T : new()
	{
		if (tmp == null)
		{
			tmp = new object[_cols.Count];
		}
		for (int i = 0; i < _cols.Count; i++)
		{
			tmp[i] = _cols[i, line].ToObject();
		}
		if (arr == null)
		{
			arr = new object[subs.Length];
		}
		for (int j = 0; j < subs.Length; j++)
		{
			int num = subs[j];
			if (num >= 0)
			{
				arr[j] = tmp[num];
			}
		}
		if (t == null)
		{
			t = XData<T>.Create(arr);
		}
		else
		{
			XData<T>.Fill(t, arr);
		}
		return t;
	}

	public string[] ToStrings()
	{
		if (_cols.Count <= 0)
		{
			return new string[0];
		}
		return ToStrings(_cols[0].Name);
	}

	public string[] ToStrings(params string[] ns)
	{
		return ToStringList(ns).ToArray();
	}

	public List<string> ToStringList(params string[] ns)
	{
		List<string> list = new List<string>();
		foreach (string name in ns)
		{
			RecordColumn recordColumn = this[name];
			if (recordColumn == null)
			{
				continue;
			}
			for (int j = 0; j < Count; j++)
			{
				string text = recordColumn[j];
				if (!string.IsNullOrEmpty(text))
				{
					int num = list.BinarySearch(text);
					if (num < 0)
					{
						list.Insert(-num - 1, text);
					}
				}
			}
		}
		return list;
	}

	public int[] ToInts()
	{
		if (_cols.Count <= 0)
		{
			return new int[0];
		}
		return ToInts(_cols[0].Name);
	}

	public int[] ToInts(params string[] ns)
	{
		return ToIntList(ns).ToArray();
	}

	public List<int> ToIntList(params string[] ns)
	{
		List<int> list = new List<int>();
		foreach (string name in ns)
		{
			RecordColumn recordColumn = this[name];
			if (recordColumn == null)
			{
				continue;
			}
			for (int j = 0; j < Count; j++)
			{
				int num = recordColumn[j].ToInt();
				if (num != 0)
				{
					int num2 = list.BinarySearch(num);
					if (num2 < 0)
					{
						list.Insert(-num2 - 1, num);
					}
				}
			}
		}
		return list;
	}

	public List<long> ToLongList(params string[] ns)
	{
		List<long> list = new List<long>();
		foreach (string name in ns)
		{
			RecordColumn recordColumn = this[name];
			if (recordColumn == null)
			{
				continue;
			}
			for (int j = 0; j < Count; j++)
			{
				long num = recordColumn[j].ToLong();
				if (num != 0L)
				{
					int num2 = list.BinarySearch(num);
					if (num2 < 0)
					{
						list.Insert(-num2 - 1, num);
					}
				}
			}
		}
		return list;
	}

	public long[] ToLongs()
	{
		if (_cols.Count <= 0)
		{
			return new long[0];
		}
		return ToLongs(_cols[0].Name);
	}

	public long[] ToLongs(params string[] ns)
	{
		List<long> list = new List<long>();
		foreach (string name in ns)
		{
			RecordColumn recordColumn = this[name];
			if (recordColumn == null)
			{
				continue;
			}
			for (int j = 0; j < Count; j++)
			{
				long num = recordColumn[j].ToLong();
				if (num != 0L)
				{
					int num2 = list.BinarySearch(num);
					if (num2 < 0)
					{
						list.Insert(-num2 - 1, num);
					}
				}
			}
		}
		return list.ToArray();
	}

	public Guid[] ToGuids(params string[] ns)
	{
		List<Guid> list = new List<Guid>();
		foreach (string name in ns)
		{
			RecordColumn recordColumn = this[name];
			if (recordColumn == null)
			{
				continue;
			}
			for (int j = 0; j < Count; j++)
			{
				Guid guid = recordColumn[j].ToGuid();
				if (!(guid == Guid.Empty))
				{
					list.Add(guid);
				}
			}
		}
		return list.ToArray();
	}

	public string[] ToArray(params string[] ns)
	{
		List<string> list = new List<string>();
		foreach (string name in ns)
		{
			RecordColumn recordColumn = this[name];
			if (recordColumn == null)
			{
				continue;
			}
			for (int j = 0; j < Count; j++)
			{
				string text = recordColumn[j].ToString();
				if (text.Length != 0)
				{
					int num = list.BinarySearch(text);
					if (num < 0)
					{
						list.Insert(-num - 1, text);
					}
				}
			}
		}
		return list.ToArray();
	}

	public Record ToRecord(params string[] names)
	{
		if (names.Length == 0)
		{
			return this;
		}
		Record record = new Record();
		foreach (string name in names)
		{
			RecordColumn recordColumn = this[name];
			if (recordColumn != null)
			{
				record.Columns.Add(recordColumn.Copy()).Record = record;
			}
		}
		record.Count = Count;
		record.Cursor = 0;
		return record;
	}

	public override string ToString()
	{
		StringBuilder stringBuilder = new StringBuilder();
		stringBuilder.Append((Name.Length == 0) ? "None" : Name);
		stringBuilder.Append(" count " + Count);
		if (_maxcount > 0)
		{
			stringBuilder.Append(" max " + _maxcount);
		}
		if (_pagesize > 0)
		{
			stringBuilder.Append(" page " + Page + " per " + _pagesize);
		}
		stringBuilder.Append('\n');
		if (Count == 1)
		{
			int num = 0;
			for (int i = 0; i < _cols.Count; i++)
			{
				RecordColumn recordColumn = _cols[i];
				if (recordColumn.Name.Length > num)
				{
					num = recordColumn.Name.Length;
				}
			}
			for (int j = 0; j < _cols.Count; j++)
			{
				RecordColumn recordColumn2 = _cols[j];
				stringBuilder.Append(_cols[j].Name.PadRight(num));
				stringBuilder.Append(" | ");
				stringBuilder.Append(recordColumn2.ToString());
				stringBuilder.Append('\n');
			}
		}
		else
		{
			int[] array = new int[_cols.Count];
			string[,] array2 = new string[Count, _cols.Count];
			for (int k = 0; k < _cols.Count; k++)
			{
				RecordColumn recordColumn3 = _cols[k];
				int num2 = recordColumn3.Name.Length;
				for (int l = 0; l < Count; l++)
				{
					string text = recordColumn3[l].ToString();
					int num3 = MyText.bLength(text);
					if (num3 > 40)
					{
						text = MyText.bSubstring(text, 38) + "..";
						num3 = 40;
					}
					array2[l, k] = text;
					if (num3 > num2)
					{
						num2 = num3;
					}
				}
				array[k] = num2;
			}
			for (int m = 0; m < _cols.Count; m++)
			{
				if (m > 0)
				{
					stringBuilder.Append(" | ");
				}
				stringBuilder.Append(_cols[m].Name.PadRight(array[m]));
			}
			for (int n = 0; n < Count; n++)
			{
				stringBuilder.Append('\n');
				for (int num4 = 0; num4 < _cols.Count; num4++)
				{
					if (num4 > 0)
					{
						stringBuilder.Append(" | ");
					}
					string text2 = array2[n, num4];
					int num5 = MyText.bLength(text2);
					int num6 = array[num4];
					stringBuilder.Append(text2);
					if (num6 > num5)
					{
						stringBuilder.Append(new string(' ', num6 - num5));
					}
				}
			}
		}
		return stringBuilder.ToString();
	}

	public string ToJson(params string[] flds)
	{
		if (Count == 0)
		{
			return "[]";
		}
		using JsonWriter jsonWriter = new JsonWriter();
		if (_pagesize > 0)
		{
			if (_maxcount <= 0)
			{
				_maxcount = Count;
			}
			jsonWriter.WriteStartObject();
			jsonWriter.WriteNameValue("page", Page);
			jsonWriter.WriteNameValue("pagesize", PageSize);
			jsonWriter.WriteNameValue("maxcount", _maxcount);
			jsonWriter.WriteNameValue("maxpage", MaxPage);
			jsonWriter.WritePropertyName("data");
			WriteRecord(jsonWriter, flds);
			jsonWriter.WriteEndObject();
		}
		else
		{
			WriteRecord(jsonWriter, flds);
		}
		return jsonWriter.ToString();
	}

	private void WriteRecord(JsonWriter write, params string[] flds)
	{
		write.WriteStartArray();
		while (Read())
		{
			write.WriteStartObject();
			for (int i = 0; i < Columns.Count; i++)
			{
				RecordColumn recordColumn = Columns[i];
				if (flds != null && flds.Length != 0 && Tool.IndexOf(flds, recordColumn.Name) == -1)
				{
					continue;
				}
				write.WritePropertyName(recordColumn.Name);
				Type type = recordColumn.Type;
				if (type == typeof(string))
				{
					write.WriteValue(Get(recordColumn.Name));
					continue;
				}
				if (type == typeof(bool))
				{
					write.WriteValue(GetBool(recordColumn.Name));
					continue;
				}
				if (type == typeof(int))
				{
					write.WriteValue(GetInt(recordColumn.Name));
					continue;
				}
				if (type == typeof(long))
				{
					write.WriteValue(GetLong(recordColumn.Name));
					continue;
				}
				if (type == typeof(float))
				{
					write.WriteValue(GetFloat(recordColumn.Name));
					continue;
				}
				if (type == typeof(double))
				{
					write.WriteValue(GetDouble(recordColumn.Name));
					continue;
				}
				if (type == typeof(decimal))
				{
					write.WriteValue(GetMoney(recordColumn.Name));
					continue;
				}
				if (type == typeof(DateTime))
				{
					write.WriteValue(GetDate(recordColumn.Name).ToString("yyyy-MM-dd HH:mm:ss"));
					continue;
				}
				if (type == typeof(Guid))
				{
					write.WriteValue(Get(recordColumn.Name));
					continue;
				}
				if (type == typeof(string[]))
				{
					string[] array = recordColumn.ToStrings();
					write.WriteStartArray();
					string[] array2 = array;
					foreach (string value in array2)
					{
						write.WriteValue(value);
					}
					write.WriteEndArray();
					continue;
				}
				if (type == typeof(int[]))
				{
					int[] array3 = recordColumn.ToInts();
					write.WriteStartArray();
					int[] array4 = array3;
					foreach (int value2 in array4)
					{
						write.WriteValue(value2);
					}
					write.WriteEndArray();
					continue;
				}
				if (type == typeof(long[]))
				{
					long[] array5 = recordColumn.ToLongs();
					write.WriteStartArray();
					long[] array6 = array5;
					foreach (long value3 in array6)
					{
						write.WriteValue(value3);
					}
					write.WriteEndArray();
					continue;
				}
				if (type == typeof(double[]) || type == typeof(decimal[]))
				{
					decimal[] array7 = recordColumn.ToDecimals();
					write.WriteStartArray();
					decimal[] array8 = array7;
					foreach (decimal value4 in array8)
					{
						write.WriteValue(value4);
					}
					write.WriteEndArray();
					continue;
				}
				throw new Exception("未知数据类型" + type.Name);
			}
			write.WriteEndObject();
		}
		write.WriteEndArray();
	}

	public string ToJsonArray()
	{
		string text = "T";
		if (Name.Length > 0)
		{
			text = text + "." + Name;
		}
		if (Count == 0)
		{
			return text + "=[];";
		}
		StringBuilder stringBuilder = new StringBuilder(text + "=");
		StringBuilder stringBuilder2 = null;
		stringBuilder.Append("[");
		for (int i = 0; i < Count; i++)
		{
			stringBuilder2 = new StringBuilder();
			foreach (RecordColumn column in Columns)
			{
				if (stringBuilder2.Length > 0)
				{
					stringBuilder2.Append(',');
				}
				stringBuilder2.Append(column[i].ToSafeString());
			}
			if (i > 0)
			{
				stringBuilder.Append(',');
			}
			stringBuilder.Append("[" + stringBuilder2.ToString() + "]");
		}
		stringBuilder.Append("];");
		return stringBuilder.ToString();
	}

	public string ToPair(string key, string value)
	{
		string text = "T";
		if (Name.Length > 0)
		{
			text = text + "." + Name;
		}
		if (Count == 0)
		{
			return text + "={};";
		}
		StringBuilder stringBuilder = new StringBuilder();
		RecordColumn recordColumn = this[key];
		RecordColumn recordColumn2 = this[value];
		for (int i = 0; i < Count; i++)
		{
			if (stringBuilder.Length > 0)
			{
				stringBuilder.Append(',');
			}
			stringBuilder.Append(recordColumn[i].ToString() + ":" + recordColumn2[i].ToSafeString());
		}
		return text + "={" + stringBuilder.ToString() + "};";
	}

	public string ToCsv()
	{
		if (Count == 0)
		{
			return "";
		}
		StringBuilder stringBuilder = new StringBuilder();
		for (int i = 0; i < Count; i++)
		{
			if (i > 0)
			{
				stringBuilder.Append(",");
			}
			foreach (RecordColumn column in Columns)
			{
				stringBuilder.Append("\"" + column[i].ToString().Replace("\"", "\"\"") + "\"");
			}
			stringBuilder.Append("\r\n");
		}
		return stringBuilder.ToString();
	}

	public NameValueCollection ToQueryString()
	{
		NameValueCollection nameValueCollection = HttpUtility.ParseQueryString(string.Empty);
		while (Read())
		{
			foreach (RecordColumn column in Columns)
			{
				nameValueCollection.Add(column.Name, Get(column.Name));
			}
		}
		return nameValueCollection;
	}

	public DataTable ToTable()
	{
		DataTable dataTable = new DataTable(Name);
		foreach (RecordColumn column in Columns)
		{
			dataTable.Columns.Add(column.Name, column.Type);
		}
		for (int i = 0; i < Count; i++)
		{
			DataRow dataRow = dataTable.NewRow();
			foreach (RecordColumn column2 in Columns)
			{
				dataRow[column2.Name] = column2[i].ToObject();
			}
			dataTable.Rows.Add(dataRow);
		}
		return dataTable;
	}

	public DataSet ToDataSet()
	{
		return new DataSet
		{
			Tables = { ToTable() }
		};
	}

	public void WriteXml(XmlWriter w)
	{
		WriteXml(w, close: true);
	}

	public void WriteXml(XmlWriter w, bool close)
	{
		if (Name.Length > 0)
		{
			w.WriteStartElement(Name);
		}
		else
		{
			w.WriteStartElement("none");
		}
		if (_pagesize > 0)
		{
			if (Page > 1)
			{
				w.WriteAttributeString("page", Page.ToString());
			}
			if (_pagesize > 0)
			{
				w.WriteAttributeString("pagesize", _pagesize.ToString());
			}
			if (_maxcount > 0)
			{
				w.WriteAttributeString("maxcount", _maxcount.ToString());
			}
		}
		for (int i = 0; i < Count; i++)
		{
			w.WriteStartElement("row");
			foreach (RecordColumn column in Columns)
			{
				w.WriteAttributeString(column.Name, column[i].ToString());
			}
			w.WriteEndElement();
		}
		if (close)
		{
			w.Close();
		}
	}

	public string ToXml()
	{
		StringBuilder stringBuilder = new StringBuilder();
		for (int i = 0; i < Count; i++)
		{
			stringBuilder.Append("<row");
			foreach (RecordColumn column in Columns)
			{
				stringBuilder.Append(" " + column.Name + "=\"" + column[i].ToSafeXml() + "\"");
			}
			stringBuilder.Append(" />");
		}
		return stringBuilder.ToString();
	}

	public XmlSchema GetSchema()
	{
		return null;
	}

	public void ReadXml(XmlReader r)
	{
		while (r.Read())
		{
			switch (r.NodeType)
			{
			case XmlNodeType.Element:
				if (r.Name != "row")
				{
					if (r.Name != "none" && r.Name.Length > 0)
					{
						Name = r.Name;
					}
					while (r.MoveToNextAttribute())
					{
						if (r.Name.Equals("page"))
						{
							Page = Valid.ToInt(r.Value);
						}
						else if (r.Name.Equals("pagesize"))
						{
							PageSize = Valid.ToInt(r.Value);
						}
						else if (r.Name.Equals("maxcount"))
						{
							_maxcount = Valid.ToInt(r.Value);
						}
					}
				}
				else if (r.AttributeCount > 0)
				{
					AddRow();
					while (r.MoveToNextAttribute())
					{
						Columns.Add(r.Name).Set(r.Value);
					}
				}
				break;
			case XmlNodeType.EndElement:
				Cursor = 0;
				return;
			}
		}
	}

	public void ReadXml(string XmlStr)
	{
		RegexOptions options = RegexOptions.IgnoreCase | RegexOptions.Multiline | RegexOptions.Compiled | RegexOptions.IgnorePatternWhitespace;
		foreach (Match item in System.Text.RegularExpressions.Regex.Matches(XmlStr, "<row[^>]*?\\/>", options))
		{
			AddRow();
			foreach (Match item2 in System.Text.RegularExpressions.Regex.Matches(item.Value, "(\\w+)\\s*=\"([^\"]*)\"", options))
			{
				string value = item2.Groups[1].Value;
				string value2 = Xml.UnEscape(item2.Groups[2].Value);
				Columns.Add(value).Set(value2);
			}
		}
		Cursor = 0;
	}

	public Record Clone()
	{
		Record record = new Record(Name, RowSize);
		record.Name = Name;
		record.Page = 1;
		record._pagesize = _pagesize;
		record._maxcount = 0;
		foreach (RecordColumn column in Columns)
		{
			record.Columns.Add(column.Name, column.TypeCode, column.IsArray);
		}
		return record;
	}

	public Record Copy()
	{
		return Copy(backward: false);
	}

	public Record Copy(bool backward)
	{
		Record record = new Record(Name, RowSize);
		Copy(record, backward);
		return record;
	}

	public void Copy(Record re)
	{
		Copy(re, backward: false);
	}

	public void Copy(Record re, bool backward)
	{
		re.Name = Name;
		re.Page = Page;
		re._pagesize = _pagesize;
		re._maxcount = _maxcount;
		re.Columns.Clear();
		re.RowSize = RowSize;
		if (backward)
		{
			for (int num = Count - 1; num >= 0; num--)
			{
				re.AddRow();
				foreach (RecordColumn column in Columns)
				{
					re.Columns.Add(column.Name, column.TypeCode, column.IsArray).Set(column[num]);
				}
			}
		}
		else
		{
			foreach (RecordColumn column2 in Columns)
			{
				Array.Copy(column2.Data, re.Columns.Add(column2.Name, column2.TypeCode, column2.IsArray).Data, Count);
			}
		}
		re.Count = Count;
	}

	public void CopyLine(Record reTo)
	{
		reTo.AddRow();
		foreach (RecordColumn column in Columns)
		{
			reTo.Columns.Add(column.Name, column.TypeCode, column.IsArray).Set(column[Cursor]);
		}
	}

	public void Write(Stream sw)
	{
		WriteSchema(sw);
		MoveFirst();
		while (Read())
		{
			WriteLine(sw);
		}
	}

	public void WriteSchema(Stream sw)
	{
		sw.WriteByte((byte)Columns.Count);
		sw.Write(BitConverter.GetBytes(Count), 0, 4);
		sw.Write(BitConverter.GetBytes(Page), 0, 4);
		sw.Write(BitConverter.GetBytes(PageSize), 0, 4);
		sw.Write(BitConverter.GetBytes(MaxCount), 0, 4);
		Encoding uTF = Encoding.UTF8;
		if (Name.Length == 0)
		{
			sw.WriteByte(0);
		}
		else
		{
			byte[] bytes = uTF.GetBytes(Name);
			sw.WriteByte((byte)bytes.Length);
			sw.Write(bytes, 0, bytes.Length);
		}
		foreach (RecordColumn column in Columns)
		{
			byte[] bytes2 = uTF.GetBytes(column.Name);
			sw.WriteByte((byte)bytes2.Length);
			sw.Write(bytes2, 0, bytes2.Length);
			byte b = (byte)column.TypeCode;
			if (column.IsArray)
			{
				b += 100;
			}
			sw.WriteByte(b);
		}
	}

	public void WriteLine(Stream sw, long offset)
	{
		if (sw.Position != offset)
		{
			sw.Seek(offset, SeekOrigin.Begin);
		}
		WriteLine(sw, sw.Position);
	}

	public void WriteLine(Stream sw)
	{
		byte[] array = null;
		foreach (RecordColumn column in Columns)
		{
			switch (column.TypeCode)
			{
			case TypeCode.Boolean:
				array = BitConverter.GetBytes(column[Cursor].ToBool());
				break;
			case TypeCode.Byte:
				array = ((!column.IsArray) ? new byte[1] { column[Cursor].ToByte() } : Converter.ToBytes(column[Cursor].ToBytes()));
				break;
			case TypeCode.Char:
				array = BitConverter.GetBytes(column[Cursor].ToChar());
				break;
			case TypeCode.Double:
				array = BitConverter.GetBytes(column[Cursor].ToDouble());
				break;
			case TypeCode.Int16:
				array = BitConverter.GetBytes(column[Cursor].ToShort());
				break;
			case TypeCode.Int32:
				array = ((!column.IsArray) ? BitConverter.GetBytes(column[Cursor].ToInt()) : Converter.ToBytes(column[Cursor].ToInts()));
				break;
			case TypeCode.Int64:
				array = ((!column.IsArray) ? BitConverter.GetBytes(column[Cursor].ToLong()) : Converter.ToBytes(column[Cursor].ToLongs()));
				break;
			case TypeCode.Decimal:
				array = ((!column.IsArray) ? Converter.ToBytes(column[Cursor].ToDecimal()) : Converter.ToBytes(column[Cursor].ToDecimals()));
				break;
			case TypeCode.SByte:
				array = new byte[1] { Convert.ToByte(column[Cursor].ToSByte()) };
				break;
			case TypeCode.Single:
				array = BitConverter.GetBytes(column[Cursor].ToFloat());
				break;
			case TypeCode.UInt16:
				array = BitConverter.GetBytes(column[Cursor].ToUShort());
				break;
			case TypeCode.UInt32:
				array = BitConverter.GetBytes(column[Cursor].ToUInt());
				break;
			case TypeCode.UInt64:
				array = BitConverter.GetBytes(column[Cursor].ToULong());
				break;
			case TypeCode.DateTime:
				array = BitConverter.GetBytes(column[Cursor].ToDate().Ticks);
				break;
			case TypeCode.Empty:
				array = column[Cursor].ToGuid().ToByteArray();
				break;
			case TypeCode.String:
				if (column.IsArray)
				{
					array = Converter.ToBytes(column[Cursor].ToStrings());
					break;
				}
				array = Encoding.UTF8.GetBytes(column[Cursor].ToString());
				sw.Write(BitConverter.GetBytes(array.Length), 0, 4);
				break;
			case TypeCode.Object:
				array = Tool.Serialize(column[Cursor].ToObject());
				sw.Write(BitConverter.GetBytes(array.Length), 0, 4);
				break;
			default:
				continue;
			}
			if (array.Length != 0)
			{
				sw.Write(array, 0, array.Length);
			}
		}
	}

	public void Load(Stream sr)
	{
		int num = LoadSchema(sr);
		for (int i = 0; i < num; i++)
		{
			AddRow();
			if (!LoadLine(sr))
			{
				return;
			}
		}
		MoveFirst();
	}

	public int LoadSchema(Stream sr)
	{
		Encoding uTF = Encoding.UTF8;
		int num = sr.ReadByte();
		byte[] array = new byte[4];
		sr.Read(array, 0, array.Length);
		int result = BitConverter.ToInt32(array, 0);
		sr.Read(array, 0, array.Length);
		Page = BitConverter.ToInt32(array, 0);
		sr.Read(array, 0, array.Length);
		PageSize = BitConverter.ToInt32(array, 0);
		sr.Read(array, 0, array.Length);
		MaxCount = BitConverter.ToInt32(array, 0);
		int num2 = sr.ReadByte();
		if (num2 > 0)
		{
			byte[] array2 = new byte[num2];
			sr.Read(array2, 0, num2);
			Name = uTF.GetString(array2);
		}
		for (int i = 0; i < num; i++)
		{
			num2 = sr.ReadByte();
			byte[] array3 = new byte[num2];
			sr.Read(array3, 0, num2);
			string name = uTF.GetString(array3);
			int num3 = sr.ReadByte();
			bool flag = num3 > 100;
			if (flag)
			{
				num3 -= 100;
			}
			Columns.Add(new RecordColumn(this, name, (TypeCode)num3, flag));
		}
		return result;
	}

	public bool LoadLine(Stream sr, long offset)
	{
		if (sr.Position != offset)
		{
			sr.Seek(offset, SeekOrigin.Begin);
		}
		return LoadLine(sr, sr.Position);
	}

	public bool LoadLine(Stream sr)
	{
		long num = sr.Length - sr.Position;
		foreach (RecordColumn column in Columns)
		{
			switch (column.TypeCode)
			{
			case TypeCode.Boolean:
			{
				if (num < 1)
				{
					return false;
				}
				byte[] array = new byte[1];
				sr.Read(array, 0, array.Length);
				column[Cursor].Set(BitConverter.ToBoolean(array, 0));
				break;
			}
			case TypeCode.Byte:
			{
				if (num < 1)
				{
					return false;
				}
				if (column.IsArray)
				{
					column[Cursor].Set(Converter.ReadBytes(sr));
					break;
				}
				byte[] array = new byte[1];
				sr.Read(array, 0, array.Length);
				column[Cursor].Set(array[0]);
				break;
			}
			case TypeCode.Char:
			{
				if (num < 2)
				{
					return false;
				}
				byte[] array = new byte[2];
				sr.Read(array, 0, array.Length);
				column[Cursor].Set(BitConverter.ToChar(array, 0));
				break;
			}
			case TypeCode.Double:
			{
				if (num < 8)
				{
					return false;
				}
				byte[] array = new byte[8];
				sr.Read(array, 0, array.Length);
				column[Cursor].Set(BitConverter.ToDouble(array, 0));
				break;
			}
			case TypeCode.Int16:
			{
				if (num < 2)
				{
					return false;
				}
				byte[] array = new byte[2];
				sr.Read(array, 0, array.Length);
				column[Cursor].Set(BitConverter.ToInt16(array, 0));
				break;
			}
			case TypeCode.Int32:
			{
				if (num < 4)
				{
					return false;
				}
				if (column.IsArray)
				{
					column[Cursor].Set(Converter.ToInts(sr));
					break;
				}
				byte[] array = new byte[4];
				sr.Read(array, 0, array.Length);
				column[Cursor].Set(BitConverter.ToInt32(array, 0));
				break;
			}
			case TypeCode.Int64:
			{
				if (num < 8)
				{
					return false;
				}
				if (column.IsArray)
				{
					column[Cursor].Set(Converter.ToLongs(sr));
					break;
				}
				byte[] array = new byte[8];
				sr.Read(array, 0, array.Length);
				column[Cursor].Set(BitConverter.ToInt64(array, 0));
				break;
			}
			case TypeCode.Decimal:
			{
				if (num < 16)
				{
					return false;
				}
				if (column.IsArray)
				{
					column[Cursor].Set(Converter.ToDecimals(sr));
					break;
				}
				byte[] array = new byte[16];
				sr.Read(array, 0, array.Length);
				column[Cursor].Set(Converter.ToDecimal(array));
				break;
			}
			case TypeCode.SByte:
			{
				if (num < 1)
				{
					return false;
				}
				byte[] array = new byte[1];
				sr.Read(array, 0, array.Length);
				column[Cursor].Set(Convert.ToSByte(array[0]));
				break;
			}
			case TypeCode.Single:
			{
				if (num < 4)
				{
					return false;
				}
				byte[] array = new byte[4];
				sr.Read(array, 0, array.Length);
				column[Cursor].Set(BitConverter.ToSingle(array, 0));
				break;
			}
			case TypeCode.UInt16:
			{
				if (num < 2)
				{
					return false;
				}
				byte[] array = new byte[2];
				sr.Read(array, 0, array.Length);
				column[Cursor].Set(BitConverter.ToUInt16(array, 0));
				break;
			}
			case TypeCode.UInt32:
			{
				byte[] array = new byte[4];
				sr.Read(array, 0, array.Length);
				column[Cursor].Set(BitConverter.ToUInt32(array, 0));
				break;
			}
			case TypeCode.UInt64:
			{
				if (num < 8)
				{
					return false;
				}
				byte[] array = new byte[8];
				sr.Read(array, 0, array.Length);
				column[Cursor].Set(BitConverter.ToUInt64(array, 0));
				break;
			}
			case TypeCode.DateTime:
			{
				if (num < 8)
				{
					return false;
				}
				byte[] array = new byte[8];
				sr.Read(array, 0, array.Length);
				column[Cursor].Set(new DateTime(BitConverter.ToInt64(array, 0)));
				break;
			}
			case TypeCode.Empty:
			{
				if (num < 16)
				{
					return false;
				}
				byte[] array = new byte[16];
				sr.Read(array, 0, array.Length);
				column[Cursor].Set(new Guid(array));
				break;
			}
			case TypeCode.String:
			{
				if (column.IsArray)
				{
					column[Cursor].Set(Converter.ToStrings(sr));
					break;
				}
				if (num < 4)
				{
					return false;
				}
				int num3 = Converter.ReadInt(sr);
				num -= 4;
				if (num3 < 0 || num3 > num)
				{
					throw new Exception(column.Name + " 长度 " + num3 + " 超过 " + num);
				}
				if (num3 == 0)
				{
					column[Cursor].Set("");
					break;
				}
				byte[] array = new byte[num3];
				sr.Read(array, 0, num3);
				column[Cursor].Set(Encoding.UTF8.GetString(array));
				break;
			}
			default:
			{
				int num2 = Converter.ReadInt(sr);
				if (num2 <= 0 || num2 > num)
				{
					throw new Exception(column.Name + " 长度 " + num2 + " 超过 " + num);
				}
				byte[] array = new byte[num2];
				sr.Read(array, 0, num2);
				column[Cursor].Set(Tool.Deserialize(array));
				break;
			}
			}
		}
		return true;
	}

	public void GetSchema(DbDataReader reader)
	{
		int visibleFieldCount = reader.VisibleFieldCount;
		for (int i = 0; i < visibleFieldCount; i++)
		{
			Columns.Add(reader.GetName(i), reader.GetFieldType(i));
		}
	}

	public void SortAsc(params string[] flds)
	{
		Sort(0, Count - 1, flds, forward: true);
	}

	public void SortDesc(params string[] flds)
	{
		Sort(0, Count - 1, flds, forward: false);
	}

	private void Sort(int left, int right, string[] flds, bool forward)
	{
		if (Count == 0)
		{
			return;
		}
		do
		{
			int i = left;
			int num = right;
			int num2 = i + (num - i >> 1);
			Swap(i, num2, flds, forward);
			Swap(i, num, flds, forward);
			Swap(num2, num, flds, forward);
			do
			{
				if (forward)
				{
					for (; Compare(i, num2, flds) < 0; i++)
					{
					}
					while (Compare(num2, num, flds) < 0)
					{
						num--;
					}
				}
				else
				{
					for (; Compare(i, num2, flds) > 0; i++)
					{
					}
					while (Compare(num2, num, flds) > 0)
					{
						num--;
					}
				}
				if (i > num)
				{
					break;
				}
				if (i < num)
				{
					Swap(i, num);
				}
				i++;
				num--;
			}
			while (i <= num);
			if (num - left <= right - i)
			{
				if (left < num)
				{
					Sort(left, num, flds, forward);
				}
				left = i;
			}
			else
			{
				if (i < right)
				{
					Sort(i, right, flds, forward);
				}
				right = num;
			}
		}
		while (left < right);
	}

	private int Compare(int a, int b, string[] flds)
	{
		int num = 0;
		for (int i = 0; i < flds.Length; i++)
		{
			num = this[flds[i]].Compare(a, b);
			if (num != 0)
			{
				return num;
			}
		}
		return num;
	}

	private void Swap(int a, int b, string[] flds, bool forward)
	{
		if (a == b)
		{
			return;
		}
		int num = Compare(a, b, flds);
		if (forward)
		{
			if (num <= 0)
			{
				return;
			}
			{
				foreach (RecordColumn column in Columns)
				{
					column.Swap(a, b);
				}
				return;
			}
		}
		if (num >= 0)
		{
			return;
		}
		foreach (RecordColumn column2 in Columns)
		{
			column2.Swap(a, b);
		}
	}

	private void Swap(int a, int b)
	{
		foreach (RecordColumn column in Columns)
		{
			column.Swap(a, b);
		}
	}
}
