using System;
using System.Collections.Generic;
using System.IO;
using System.Linq.Expressions;
using System.Reflection;

namespace ZYing;

public class Para
{
	public int page;

	public int pagesize;

	public int maxcount;

	public int top;

	public long id;

	public long pid;

	public int userid;

	public int sectid;

	public int loginid;

	public int from;

	public int aid;

	public int role;

	public int stat;

	public int type;

	public int flag;

	public int grade;

	public int level;

	public int result;

	public int term;

	public int rule;

	public int what;

	public int localid;

	public long local;

	public long kind;

	public long feed;

	public bool ban;

	public bool ignore;

	public bool done;

	public bool select;

	public decimal low;

	public decimal high;

	public string start;

	public string end;

	public string order;

	public string area;

	public string word;

	public string key;

	public string levels;

	public string grades;

	public string extra;

	public string abbr;

	public string lang;

	public string cur;

	public int min;

	public int max;

	public int[] accs;

	public int[] logins;

	public int[] sites;

	public long[] ids;

	public string[] values;

	public string[] keys;

	private static readonly Func<Stream, Para, int> _cacheWrite;

	static Para()
	{
		Type typeFromHandle = typeof(Para);
		ParameterExpression parameterExpression = Expression.Parameter(typeFromHandle, "p");
		List<Expression> list = new List<Expression>();
		ParameterExpression parameterExpression2 = Expression.Parameter(typeof(Stream), "sr");
		ParameterExpression parameterExpression3 = Expression.Parameter(typeof(int), "ret");
		list.Add(Expression.Assign(parameterExpression3, Expression.Constant(0)));
		FieldInfo[] fields = typeFromHandle.GetFields();
		foreach (FieldInfo fieldInfo in fields)
		{
			MemberExpression memberExpression = Expression.Field(parameterExpression, fieldInfo);
			MethodInfo method = typeof(Tool).GetMethod("Write", new Type[2]
			{
				typeof(Stream),
				fieldInfo.FieldType
			});
			MethodCallExpression item = Expression.Call(null, method, new Expression[2] { parameterExpression2, memberExpression });
			list.Add(item);
		}
		LabelTarget target = Expression.Label(typeof(int));
		GotoExpression item2 = Expression.Return(target, parameterExpression3);
		LabelExpression item3 = Expression.Label(target, parameterExpression3);
		list.Add(item2);
		list.Add(item3);
		_cacheWrite = Expression.Lambda<Func<Stream, Para, int>>(Expression.Block(new ParameterExpression[1] { parameterExpression3 }, list), new ParameterExpression[2] { parameterExpression2, parameterExpression }).Compile();
	}

	public static byte[] ToArray(Para p)
	{
		using MemoryStream memoryStream = new MemoryStream();
		_cacheWrite(memoryStream, p);
		return memoryStream.ToArray();
	}
}
