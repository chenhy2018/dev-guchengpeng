#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "curl.h"
#include "curl_req.h"

//#define URL_REQ "https://example.com/streams/pub/123?id=123"
#define URL_REQ "http://101.71.78.175:16301/streams/pub/test001"

static char *UrlRtmp = NULL;
void SetUrlRtmp(const char* url, const int len)
{
	if (NULL != url)
	{
		UrlRtmp = (char *)malloc(len +1);
		memcpy(UrlRtmp, url, len);
		UrlRtmp[len] = '\0';
	}

	return;
}
int GetUrlRtmp(char *url)
{
	if (NULL != url)
	{	
		if (NULL != UrlRtmp)
		{
			memcpy(url, UrlRtmp, 256);
		}
		else 
		{
			curl_ReqUrl();
			return  CURL_FAIL;
		}
	}
	if (NULL != UrlRtmp)
	{
		free(UrlRtmp);
		return CURL_SUCCEED;
	}
	
	return  CURL_FAIL;
}
/**
 * function : GetURL
 * descript : receive  the URL 
 * 
 */
size_t  curl_ReqUrlCallback(char *ptr, size_t size, size_t nmemb, void *userdata)
{
	printf("back url :%s\n", (char *)ptr);
	
	SetUrlRtmp((char *)ptr, size * nmemb);
	return size * nmemb;
}

/**
 * function : curl_ReqUrl
 * des		: request the URL RTMP push to 
 * ret val	:
 */
//int curl_ReqUrl(char *url)
int curl_ReqUrl()
{	
	CURL *pCurl;
	CURLcode res;
	
	//if (NULL != url)
	//{
		printf("request url...\n");
		curl_global_init(CURL_GLOBAL_DEFAULT);	
		pCurl = curl_easy_init();
		if (pCurl)
		{
			curl_easy_setopt(pCurl, CURLOPT_URL, URL_REQ);
			curl_easy_setopt(pCurl, CURLOPT_WRITEFUNCTION, &curl_ReqUrlCallback);
			curl_easy_setopt(pCurl, CURLOPT_TIMEOUT, 3L);
			curl_easy_setopt(pCurl, CURLOPT_CONNECTTIMEOUT, 10L);
			res = curl_easy_perform(pCurl);
			curl_easy_cleanup(pCurl);
			curl_global_cleanup();
			return CURL_SUCCEED;
		}
	//}
	//else 
	//{
		//return CURL_FAIL;
	//}
}


