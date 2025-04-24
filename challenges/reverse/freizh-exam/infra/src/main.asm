;; Small chall to understand syscalls, /proc/self/* and xor;
;;
;; Utils;
;;
;; import re;[int.from_bytes(p, 'little') for p in re.findall(b'..',b'/proc/self/cmdline\x00\x00')[::-1]]
;; len(b'/proc/self/cmdline\x00\x00')
;;
;; ",".join([f"'{c}'" for c in "Well done, you are ready to move forward."])
;; ",".join([f"'{c}'" for c in "The greatest teacher is failure."])
;;
;; WARN: no spaces in flag!!! (__spaces__ are null-char in cmdline);
;; f="plz_f0llow_CTFERIO_on_github";",".join([str(len(f)+1)]+[f"'{c}'" for c in f]+[str(0)])
;;

;; ---------------------------------------------------;
;; def macro; 

%define STDOUT      1

%define FLAG        29,'p','l','z','_','f','0','l','l','o','w','_','C','T','F','E','R','I','O','_','o','n','_','g','i','t','h','u','b',0
%define SIZEOF_BUF  64

%define SELF_CMDLINE    0, 25966, 26988, 25709, 25391, 26220, 25971, 12131, 28530, 28719
%define SIZEOF_CMDLINE  20

;; ---------------------------------------------------;
;; fn macros; 

%macro __open 1
;; open syscall wrapper;
;; %1 - path;
;;
    mov     rax,    0x02
    mov     rdi,    %1
    xor     rsi,    rsi
    xor     rdx,    rdx

    syscall
%endmacro

%macro __read 3
;; read syscall wrapper;
;; %1 - fd;
;; %2 - buf;
;; %3 - sizeof buf;
;;
    xor     rax,    rax
    mov     rdi,    %1
    lea     rsi,    %2
    mov     rdx,    %3

    syscall
%endmacro

%macro __write 3
;; read syscall wrapper;
;; %1 - fd;
;; %2 - buf;
;; %3 - sizeof buf;
;;
    mov     rax,    1
    mov     rdi,    %1
    mov     rsi,    %2
    mov     rdx,    %3

    syscall
%endmacro

%macro __stack_allocation 1-*
;; stack_allocation macro;
;; Just push each args on stack;
;; arg max is 0xffff (2 bytes);
;;
    %rep    %0

        push  word  %1

    %rotate 1
    %endrep
%endmacro

%macro __flag_validation 1-*
;; macro for flag_validation;
;; %1   - sizeof flag;
;; %2-* - each char to check;
;;
;; user const char* form user must be in rdi;
;; user size in rax;
;;
    cmp     rax,    %1                              ; check size;
    jnz     .looser                                 ; jmp if not equal;
    
    xor     rcx,    rcx                             ; reset counter;

    %rotate 1                                       ; skip first arg (size);
    %rep (%0 - 1)
        mov     al,     [rdi+rcx]                   ; retrieve one char;
        
        xor     al,     %1
        test    al,     al                          ; check if this is a match;
        jnz     .looser                             ; jmp if not;

        inc     rcx                                 ; next char;
    %rotate 1
    %endrep
%endmacro

;; ---------------------------------------------------;
;; start; 

section .text
global _start

_start:
    __stack_allocation SELF_CMDLINE
    __open  rsp                             ; open cmdline (should return an fd);
    add     rsp,    SIZEOF_CMDLINE          ; free space;

    cmp     rax,    0                       ; check if -1 (error);
    jle     .error                          ; jmp if error;
    
    sub     rsp,    SIZEOF_BUF              ; allocation on stack;
    push    rax                             ; push fd onto stack (after allocation);
    
    __read  [rsp], [rsp+8], SIZEOF_BUF      ; __read from fd, into buf also on stack;
    
    pop     rdi                             ; fd in rdi (in place to close);
    push    rax                             ; save __read result (n of char readed);

    mov     rax,    3
    syscall                                 ; close(fd);

    lea     rdi,    [rsp+8]                 ; ptr to buf;
    mov     rcx,    SIZEOF_BUF              ; max iteration;
    xor     al,     al                      ; char to search (null);
    repne   scasb                           ; will inc rdi til null-char;

    lea     rax,    [rsp+8]                 ; ptr to buf;
    xchg    rax,    rdi                     ; rax is now the ptr incremented;

    sub     rax,    rdi                     ; rax now contain how many bytes to get to argv[1];
    sub     [rsp],  rax                     ; decrement readed size (now this is the sizeof argv[1..]);
    add     rdi,    rax                     ; increment ptr so now it point to argv[1];

    pop     rax                             ; retrieve sizeof argv[1..];

    __flag_validation FLAG

.winner:
    add     rsp,    SIZEOF_BUF

    jmp     .winner_msg
.winner_alloc:
    db      'W','e','l','l',' ','d','o','n','e',',',' ','y','o','u',' ','a','r','e',' ','r','e','a','d','y',' ','t','o',' ','m','o','v','e',' ','f','o','r','w','a','r','d','.',0x0A,0
    SIZEOF_WINNER equ $-.winner_alloc

.winner_msg:
    __write STDOUT, .winner_alloc, SIZEOF_WINNER

    mov     rax,    0x3C                    ; exit(
    xor     rdi,    rdi                     ;   0
    syscall                                 ; );

    ret

.looser:
    add     rsp,    SIZEOF_BUF

    jmp     .looser_msg
.looser_alloc:
    db      'T','h','e',' ','g','r','e','a','t','e','s','t',' ','t','e','a','c','h','e','r',' ','i','s',' ','f','a','i','l','u','r','e','.',0x0A,0
    SIZEOF_LOOSER equ $-.looser_alloc

.looser_msg:
    __write STDOUT, .looser_alloc, SIZEOF_LOOSER

.error:
    mov     rax,    0x3C                    ; exit(
    xor     rdi,    rdi
    not     rdi                             ;   -1
    syscall                                 ; );

    ret


;; ---------------------------------------------------;
;; eof; 
