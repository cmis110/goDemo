package main

import ()

func main() {

}

/*
国际化与本地化
为了适应经济的全球一体化 作为开发者 我们需要开发出 支持多国语言 国际化的web应用
即同样的页面在不同的语言环境下需要显示不同的效果 也就是说 应用程序在运行时能够根据请求所
来自的地域与语言的不同而显示不同的用户界面 这样 当需要在应用程序中添加对新语言的支持时
无需改动应用程序的代码 只需要增加语言包即可实现

国际化与本地化Internationalization and localization 通常用i18n和L10N表示
国际化是将针对某个地区设计的程序进行重构 使它能够在更多的地区使用 本地化是指在一个面向国际化的程序中
增加对新地区的支持
目前 GO语言的标准包 没有提供对i18n的支持 但有一些比较简单的第三方实现 这一章 我们将实现一个go-i18n库
用来支持Go语言的i18n

所谓国际化 就是根据特定的local信息 提取与之相应的字符串或其它一些东西 比如时间和货币的格式 等等
这涉及到三个问题
1 如何确定local
2 如何保存与local相关的字符串或其它信息
3 如何根据local提取字符串和其他相应的信息

在第一小节里 我们将介绍如何设置正确的local以便让访问站点的用户能够获得与其语言相应的页面
第二小节将介绍如何处理或者存储字符串 货币 时间日期等与local相关的信息
第三小节将介绍如何实现国际化站点 即如何根据不同的local返回不同合适的内容
通过这三小节的学习 我们将获得一个完整的 i18n方案
/

/*
10.1 设置默认地区
什么是Local
local是一组描述世界上某一特定区域文本格式和语言习惯的设置的集合
local名通常由三个部分组成 第一部分是一个强制性的 表示语言的缩写 例如en 英文 zh中文
第二部分 跟在一个下划线之后 是一个可选国家说明符 用于区分讲同一种语言的不同国家 en_US en_UK
最后一部分 跟在一个句点之后 是可选的字符集说明 例如 zh_CN.gb2312 表示中国使用gb2312字符集

GO语言默认采用utf-8编码 所以我们实现i18n时不考虑第三部分 读者可以看到这些地区名的命名规范
对于BSD等系统 没有local命令 但是地区信息存储在/usr/share/local中

设置Locale
有了上面对于locale的定义 那么我们就需要根据用户的信息 访问信息 个人信息 访问域名 等来设置与之相关的locale
我们可以通过如下几种方式来设置用户的locale

通过域名设置locale
设置Locale的办法这一就是在应用运行的时候采用域名分级的方式 例如我们采用 www.jianling.com当做我们的英文站
把www.jianling.cn当做中文站 这样通过应用里面设置域名 和相应的locale的对应关系 就可以设置好地区
这样处理的好处
通过URL就可以很明显的识别
用户可以通过域名很直观的知道将访问哪种语言的站点
在CO程序中实现非常的简单方便 通过一个map就可以实现
有利于搜索引擎抓取 能够提高站点的seo

我们可以通过下面的代码来实现域名对应locale
if r.Host=="www.jianling.com"{
	i18n.SetLocale("en")
}else if r.Host=="www.jianling.cn"{
	i18n.SetLocale("zh-CN")
}else if r.Host=="www.jianling.tw"{
	i18n.SetLocale("zh-TW")
}

当然除了整域名设置地区之外 我们还可以通过子域名来设置地区 例如 en.jianling.com 表示英文站点 cn.jianling,com表示中文站点
实现代码如下

prefix :=strings.Split(r.Host,".")
if prefix[0] =="en"{
	i18n.SetLocale("en")
}else if prefix[0]	=="cn"{
	i18n.SetLocale("zh-CN")
}else if prefix[0] =="tw"{
	i18n.SetLocale("zh-TW")
}
通过域名设置Locale有如上所示的优点 但是我们一般开发web应用的时候不会采用这种方式 因为首先域名成本比较高
开发一个Locale就需要一个域名 而且往往统一的域名不一定能申请到 其次我们不愿意为每个站点去本地化一个配置
而更多的采用url后面带参数的方式 请看下面的介绍

从域名参数设置Locale
目前最常用的设置Locale的方式是在URL里面带上参数 例如www.jianling.com/hello?local=2h或者www.jianling.com/zh/hello
这样我们就可以设置地区 i18n.SetLocale(params["locale"])
这种设置方式几乎拥有前面讲的通过域名设置Locale的所有优点 它采用RESTful的方式 使得我们不需要增加额外的方法来处理
但是这种方式需要在每一个link里面增加相应的参数locale 这也许有点复杂而且有时候 相当的繁琐
不过我们可以写一个通用的函数url 让所有的link地址都通过这个函数来生成 然后在这个函数里面增加
locale=params["locale"]参数来缓解一下
也许我们希望URL地址看上去更加RESTful一点 例如www.jianling.com/en/books www.jianling.com/zh/books
这种方式的URL更加有利于SEO 而且对于用户也比较友好 能够通过URL直观的知道访问的站点
那么这样的URL地址可以通过router来获取locale 参考REST小节里面介绍的router插件实现
mux.Gey("/:locale/books",listbook)

从客户端设置地区
在一些特殊的情况下 我们需要根据客户端的信息 而不是通过URL来设置Locale 这些信息可能来自于客户端设置的喜好语言
用户的IP地址 用户在注册时填写的所在地等 这种方式比较适合web为基础的应用
Accept-Language
客户端请求的时候 在HTTP头信息里面有accept-language 一般的客户端都会设置该信息
下面是GO语言实现的一个简单的根据Accept-Language 实现设置地区的代码
AL:=r.Header.Get("Accept-Language")
if AL=="en"{
 i18n.SetLocale("en")
} else if AL == "zh-CN" {
    i18n.SetLocale("zh-CN")
} else if AL == "zh-TW" {
    i18n.SetLocale("zh-TW")
}
当然在实际应用中，可能需要更加严格的判断来进行设置地区

IP地址

另一种根据客户端来设定地区就是用户访问的IP，我们根据相应的IP库，对应访问的IP到地区，目前全球比较常用的就是GeoIP Lite Country这个库。这种设置地区的机制非常简单，我们只需要根据IP数据库查询用户的IP然后返回国家地区，根据返回的结果设置对应的地区。

用户profile

当然你也可以让用户根据你提供的下拉菜单或者别的什么方式的设置相应的locale，然后我们将用户输入的信息，保存到与它帐号相关的profile中，当用户再次登陆的时候把这个设置复写到locale设置中，这样就可以保证该用户每次访问都是基于自己先前设置的locale来获得页面。

总结
通过上面的介绍可知 设置locale可以有很多种方式 我们应该根据需求的不同来选择 不同的设置locale的方法
让用户能以它最熟悉的方式 获得我们提供的服务 提高应用的用户友好性



















*/
