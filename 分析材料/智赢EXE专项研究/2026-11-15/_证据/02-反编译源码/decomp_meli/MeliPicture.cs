using System.Collections.Generic;
using System.Linq;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;

namespace ZYing.Meli.Model;

public class MeliPicture
{
	public class Crop
	{
		public bool validated { get; set; }

		public object x_offset { get; set; }

		public object x_size { get; set; }

		public object y_offset { get; set; }

		public object y_size { get; set; }

		public Crop()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}
	}

	public class Variation
	{
		public string size { get; set; }

		public string secure_url { get; set; }

		public string url { get; set; }

		public Variation()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}
	}

	private int _width;

	private int _height;

	public string id { get; set; }

	public string max_size { get; set; }

	public string dominant_color { get; set; }

	public string hash { get; set; }

	public List<Variation> variations { get; set; }

	public string status { get; set; }

	public string origin { get; set; }

	public string SecureUrl => variations?.FirstOrDefault((Variation x) => x.size == max_size)?.secure_url;

	public int Width
	{
		get
		{
			if (_width > 0)
			{
				return _width;
			}
			_width = GetSize().width;
			return _width;
		}
	}

	public int Height
	{
		get
		{
			if (_height > 0)
			{
				return _height;
			}
			_height = GetSize().height;
			return _height;
		}
	}

	private (int width, int height) GetSize()
	{
		int[] array = Split.Ints(max_size, 'x');
		if (array.Length != 2)
		{
			return default;
		}
		_width = array[0];
		_height = array[1];
		return (width: _width, height: _height);
	}

	public MeliPicture()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		variations = new List<Variation>();
		base..ctor();
	}
}
