	AREA codigo, CODE, READONLY
	PRESERVE8
	EXPORT conecta_K_buscar_alineamiento_arm

conecta_K_buscar_alineamiento_arm
	mov ip, sp
	stmdb r13!, {r4-r5, fp, lr}
	sub fp, ip, #4
	; comprobamos si la fila es v?lida
	cmp r1, #1
	blt fin ; si fila < 1
	cmp r1, #6
	bgt fin ; si fila > 6
	; comprobamos si la columna es v?lida
	cmp r2, #1
	blt fin ; si columna < 1
	cmp r2, #7
	bgt fin ; si columna > 7
	; si la celda se encuentra en el tablero, obtenemos su valor
	add r4, r0, r1, LSL #3
	ldrb r4, [r4, r2] ; r4 - cuadricula[fila][columna]
	; comprobamos si la celda es vac?a
	and r5, r4, #4
	cmp r5, #0
	beq fin ; si celda es vac?a
	; comprobamos si el color de la celda es el color de la l?nea que estamos comprobando
	and r4, r4, #3
	cmp r4, r3
	bne fin ; si los colores no coinciden
	; preparamos los par?metros para llamar a la funci?n
	ldrsb r4, [fp, #4] ; r4 - deltas_fila[i]
	add r1, r1, r4 ; r1 - fila + deltas_fila[i]
	ldrsb r5, [fp, #8] ; r5 - deltas_columna[i]
	add r2, r2, r5 ; r2 - columna + deltas_columna[i]
	stmdb r13!, {r4-r5} ; se apilan los deltas para la siguiente llamada recursiva
	bl conecta_K_buscar_alineamiento_arm
	add sp, sp, #8
	add r0, r0, #1 ; r0 - r0 + 1 (return 1 + resultado funci?n)
	ldmia r13!, {r4-r5, fp, lr}
	bx r14
fin
	mov r0, #0 ; r0 - 0 (return 0)
	ldmia r13!, {r4-r5, fp, lr}
	bx r14

END