using System;
using System.Buffers;
using System.Collections;
using System.Collections.Generic;
using System.Data;
using System.IO;
using System.Runtime.InteropServices;
using System.Runtime.Serialization;
using System.Xml;
using System.Xml.Schema;
using System.Xml.Serialization;
using ZYing.Web;

namespace ZYing.Data;

[Serializable]
public class RecordSet : ISerializable, IXmlSerializable, IEnumerable
{
	private List<Record> _list = new List<Record>();

	public int Count => _list.Count;

	public Record Root => Create("root");

	public Record this[string name]
	{
		get
		{
			for (int i = 0; i < _list.Count; i++)
			{
				Record record = _list[i];
				if (string.Compare(record.Name, name, ignoreCase: true) == 0)
				{
					return record;
				}
			}
			return new Record();
		}
		set
		{
			Add(name, value);
		}
	}

	public Record this[int index] => _list[index];

	public static RecordSet From(ref ReadOnlySequence<byte> buf)
	{
		RecordSet recordSet = new RecordSet();
		if (buf.Length == 0L)
		{
			return recordSet;
		}
		using MemoryStream memoryStream = Pool.Memory.GetStream();
		foreach (ReadOnlyMemory<byte> item in buf)
		{
			if (MemoryMarshal.TryGetArray(item, out var segment))
			{
				memoryStream.Write(segment.Array, segment.Offset, segment.Count);
				continue;
			}
			ReadOnlySpan<byte> span = item.Span;
			for (int i = 0; i < span.Length; i++)
			{
				memoryStream.WriteByte(span[i]);
			}
		}
		memoryStream.Seek(0L, SeekOrigin.Begin);
		recordSet.Load(memoryStream);
		return recordSet;
	}

	public static RecordSet From(byte[] input)
	{
		RecordSet recordSet = new RecordSet();
		if (input == null || input.Length == 0)
		{
			return recordSet;
		}
		using MemoryStream sr = Pool.Memory.GetStream(input);
		recordSet.Load(sr);
		return recordSet;
	}

	public void Load(byte[] bytes, int offset = 0)
	{
		if (offset >= bytes.Length)
		{
			return;
		}
		using MemoryStream sr = Pool.Memory.GetStream((string)null, bytes, offset, bytes.Length - offset);
		Load(sr);
	}

	public void Load(Stream sr)
	{
		int num = sr.ReadByte();
		if (num > 255)
		{
			throw new Exception("数据集数量不能超过 " + byte.MaxValue);
		}
		for (int i = 0; i < num; i++)
		{
			Record record = new Record();
			record.Load(sr);
			Add(record);
		}
	}

	public RecordSet()
	{
	}

	public RecordSet(Record re)
	{
		Add(re);
	}

	private RecordSet(SerializationInfo info, StreamingContext ctxt)
	{
		int @int = info.GetInt32("k");
		for (int i = 0; i < @int; i++)
		{
			Record re = (Record)info.GetValue(i.ToString(), typeof(Record));
			Add(re);
		}
	}

	public void GetObjectData(SerializationInfo info, StreamingContext context)
	{
		info.AddValue("k", _list.Count);
		int num = 0;
		foreach (Record item in _list)
		{
			info.AddValue(num.ToString(), item);
			num++;
		}
	}

	public void AddString(string name, object value)
	{
		if (value != null)
		{
			Root.Columns.Add(name).Set(value);
		}
	}

	public void AddBool(string name, bool value)
	{
		Root.Columns.AddBool(name).Set(value);
	}

	public void AddInt(string name, int value)
	{
		Root.Columns.AddInt(name).Set(value);
	}

	public void AddLong(string name, long value)
	{
		Root.Columns.AddLong(name).Set(value);
	}

	public void AddDouble(string name, double value)
	{
		Root.Columns.AddDouble(name).Set(value);
	}

	public Record Add(string name, Record re)
	{
		re.Name = name;
		return Add(re);
	}

	public Record Add(Record re)
	{
		if (re != null)
		{
			if (string.IsNullOrEmpty(re.Name))
			{
				re.Name = "root";
			}
			re.Cursor = 0;
			for (int i = 0; i < _list.Count; i++)
			{
				if (_list[i].Name == re.Name)
				{
					_list[i] = re;
					return re;
				}
			}
			_list.Add(re);
		}
		return re;
	}

	public void Add(RecordSet rs)
	{
		if (rs == null || rs.Count <= 0)
		{
			return;
		}
		foreach (Record item in (IEnumerable)rs)
		{
			Add(item);
		}
	}

	public bool Contains(string name)
	{
		for (int i = 0; i < _list.Count; i++)
		{
			if (string.Compare(_list[i].Name, name, ignoreCase: true) == 0)
			{
				return true;
			}
		}
		return false;
	}

	public Record Create(string name)
	{
		if (string.IsNullOrEmpty(name))
		{
			name = "root";
		}
		Record record = this[name];
		if (record == null)
		{
			record = Add(new Record(name));
		}
		return record;
	}

	public Record Delete(string name)
	{
		for (int i = 0; i < _list.Count; i++)
		{
			Record record = _list[i];
			if (string.Compare(record.Name, name, ignoreCase: true) == 0)
			{
				_list.RemoveAt(i);
				return record;
			}
		}
		return null;
	}

	public Record Rename(string old, string name)
	{
		Record record = this[old];
		if (record != null)
		{
			record.Name = name;
		}
		return record;
	}

	public void Clear()
	{
		_list.Clear();
	}

	public void Union(RecordSet list)
	{
		foreach (Record item in (IEnumerable)this)
		{
			if (list.Contains(item.Name))
			{
				item.Union(list[item.Name]);
			}
		}
	}

	public byte[] ToBytes()
	{
		using MemoryStream memoryStream = Pool.Memory.GetStream();
		Write(memoryStream);
		return memoryStream.ToArray();
	}

	public void Write(Stream sw)
	{
		if (Count > 255)
		{
			throw new Exception("数据集数量不能超过 " + byte.MaxValue);
		}
		sw.WriteByte((byte)Count);
		foreach (Record item in _list)
		{
			item.Write(sw);
		}
	}

	public DataSet ToDataSet()
	{
		DataSet dataSet = new DataSet();
		foreach (Record item in (IEnumerable)this)
		{
			dataSet.Tables.Add(item.ToTable());
		}
		return dataSet;
	}

	public string ToJson()
	{
		using JsonWriter jsonWriter = new JsonWriter();
		jsonWriter.WriteStartObject();
		foreach (Record item in (IEnumerable)this)
		{
			jsonWriter.WritePropertyName(item.Name);
			jsonWriter.WriteRawValue(item.ToJson());
		}
		jsonWriter.WriteEndObject();
		return jsonWriter.ToString();
	}

	IEnumerator IEnumerable.GetEnumerator()
	{
		return _list.GetEnumerator();
	}

	XmlSchema IXmlSerializable.GetSchema()
	{
		return null;
	}

	void IXmlSerializable.ReadXml(XmlReader r)
	{
		Record record = null;
		while (r.Read())
		{
			switch (r.NodeType)
			{
			case XmlNodeType.Element:
				if (r.Name != "row")
				{
					if (r.Name != "none" && r.Name.Length > 0)
					{
						record = new Record(r.Name);
						Add(record);
					}
					while (r.MoveToNextAttribute())
					{
						if (r.Name.Equals("page"))
						{
							record.Page = Valid.ToInt(r.Value);
						}
						else if (r.Name.Equals("pagesize"))
						{
							record.PageSize = Valid.ToInt(r.Value);
						}
						else if (r.Name.Equals("maxcount"))
						{
							record.MaxCount = Valid.ToInt(r.Value);
						}
					}
				}
				else
				{
					record.AddRow();
					while (r.MoveToNextAttribute())
					{
						record.Columns.Add(r.Name).Set(r.Value);
					}
				}
				break;
			case XmlNodeType.EndElement:
				record.Cursor = 0;
				return;
			}
		}
	}

	void IXmlSerializable.WriteXml(XmlWriter writer)
	{
		foreach (Record item in (IEnumerable)this)
		{
			item.WriteXml(writer, close: false);
		}
		writer.Close();
	}
}
