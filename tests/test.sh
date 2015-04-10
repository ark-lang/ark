echo "not working yet!"
exit

for T in `echo *.aly | sort` ;
do
	if [ -z "../bin/alloyc $T -o $T.o" ]; then
		echo $T OK.
	else
		echo $T Failed.
	fi
	rm -f _gen_*
	rm -f $T.o
done
