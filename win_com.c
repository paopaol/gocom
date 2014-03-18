#include <windows.h>
#include <stdio.h>

int com_open(int __com_num)
{
	char		com[32];
	COMMTIMEOUTS 	timeout;;
	HANDLE		hf;

	sprintf(com, "COM%d", __com_num);

	hf = CreateFile(com,
			GENERIC_READ|GENERIC_WRITE,0,NULL,OPEN_EXISTING,0,NULL);
	if(hf == INVALID_HANDLE_VALUE){
		return (int)-1;
	}
	SetupComm(hf,1024 * 10,1024 * 10); 
	GetCommTimeouts((HANDLE)hf,&timeout); 
	timeout.ReadIntervalTimeout = MAXDWORD;
	timeout.ReadTotalTimeoutMultiplier = MAXDWORD;
	timeout.ReadTotalTimeoutConstant = 1000 * 1000;
	timeout.WriteTotalTimeoutMultiplier = MAXDWORD;
	timeout.WriteTotalTimeoutConstant = 1000 * 1000;
	SetCommTimeouts(hf,&timeout); 
	//PurgeComm(hf, PURGE_TXCLEAR|PURGE_RXCLEAR);
	PurgeComm((HANDLE)hf, PURGE_TXABORT | PURGE_RXABORT | PURGE_TXCLEAR | PURGE_RXCLEAR);
	return (int)hf;
}

int com_set_deadline(int __hf, int __deadline)
{
	COMMTIMEOUTS 	timeout;
	int		ret;

	ret = GetCommTimeouts((HANDLE)__hf,&timeout); 
	if(!ret){
		return -1;
	}
	timeout.ReadTotalTimeoutConstant = __deadline * 1000;
	timeout.WriteTotalTimeoutConstant = __deadline * 1000;
	ret = SetCommTimeouts((HANDLE)__hf,&timeout); 
	if(!ret){
		return -1;
	}
	return 0;
}

int com_set_read_deadline(int __hf, int __deadline)
{
	COMMTIMEOUTS 	timeout;
	int 		ret;

	ret = GetCommTimeouts((HANDLE)__hf,&timeout); 
	if(!ret){
		return -1;
	}
	timeout.ReadTotalTimeoutConstant = __deadline * 1000;
	ret = SetCommTimeouts((HANDLE)__hf,&timeout); 
	if(!ret){
		return -1;
	}
	return 0;
}

int com_set_write_deadline(int __hf, int __deadline)
{
	COMMTIMEOUTS 	timeout;;
	int 		ret;

	ret = GetCommTimeouts((HANDLE)__hf,&timeout); 
	if(!ret){
		return -1;
	}
	timeout.WriteTotalTimeoutConstant = __deadline * 1000;
	ret = SetCommTimeouts((HANDLE)__hf,&timeout); 
	if(!ret){
		return -1;
	}
	return 0;
}

int com_set(int __hf, DCB *__dcb)
{
	DCB 	dcb;
	int	ret;

	ret = GetCommState((HANDLE)__hf,&dcb);
	if(!ret){
		return -1;
	}

	dcb.BaudRate = __dcb->BaudRate; 
	dcb.ByteSize = __dcb->ByteSize; 
	dcb.Parity = __dcb->Parity; 
	dcb.StopBits = __dcb->StopBits; 

	ret = SetCommState((HANDLE)__hf, &dcb);
	if(!ret){
		return -1;
	}
	return 0;
}

int com_read(int __hf, void *__b, int __l)
{
	BOOL 		ret;
	size_t		rl;
	int 		addr = *(int *)__b;
	char 		*p = (char *)addr;
	DWORD		err;


	ret = ReadFile((HANDLE)__hf, p, __l,(PDWORD)&rl,NULL);
	if(ret == 0){	
			return -1;
	}
	//PurgeComm((HANDLE)__hf, PURGE_TXABORT | PURGE_RXABORT | PURGE_TXCLEAR | PURGE_RXCLEAR);
	return rl;
}

int com_write(int __hf, const void *__wb, int __wl)
{
	size_t 		wl;
	BOOL 		ret;
	int 		addr = *(int *)__wb;
	char 		*p = (char *)addr;

	ret = WriteFile((HANDLE)__hf, p, __wl,(PDWORD)&wl,NULL);
	//FlushFileBuffers((HANDLE)__hf);
	if(ret == 0){
		return -1;
	}
	return wl;
}

int com_close(int __hf)
{
	int 		ret;

	ret = CloseHandle((HANDLE)__hf);
	if(ret == 0){
		return -1;
	}
	return 0;
}
