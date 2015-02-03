solution "alloy"
	configurations { "Debug", "Release" }

	-- get data from shell
	LLVM_OPTIONS	= "--system-libs --libs "
	LLVM_CONFIG	= "core analysis executionengine jit interpreter native "
	LLVM_CFLAGS	= "$(" .. "llvm-config --cflags " .. LLVM_OPTIONS .. LLVM_CONFIG .. ")"
	LLVM_LFLAGS	= "$(" .. "llvm-config --ldflags " .. LLVM_OPTIONS .. LLVM_CONFIG .. ")"
	
	-- common settings
	defines { "_GNU_SOURCE", "__STDC_LIMIT_MACROS", "__STDC_CONSTANT_MACROS" }
	buildoptions { LLVM_CFLAGS, "-pthread", "-xc" }
	includedirs { "includes" }
	links { "dl", "ncurses", "z" }
	linkoptions { LLVM_LFLAGS, "-pthread", "-Wl,--as-needed -ltinfo" }

	configuration "Debug"
		flags { "Symbols", "ExtraWarnings", "FatalWarnings" }

	configuration "Release"
		buildoptions { "-march=native", "-O2" }

	-- compiler
	project "alloyc"
		kind "ConsoleApp"
		language "C++"			-- because LLVM is a bitchling
		files { "src/*.c", "src/*.h" }
