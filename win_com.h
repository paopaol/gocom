#ifndef WIN_GOCOM_H
#define WIN_GOCOM_H
#include <windows.h>
extern int com_open(int __com_num);
extern int com_set(int __hf, DCB *__dcb);
extern int com_set_deadline(int __hf, int __deadline);
extern int com_set_read_deadline(int __hf, int __deadline);
extern int com_set_write_deadline(int __hf, int __deadline);
extern int com_read(int __hf, void *__b, int __l);
extern int com_write(int __hf, const void *__wb, int __wl);
extern int com_close(int __hf);
#endif
