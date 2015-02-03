solution "alloy"
	configurations { "Debug", "Release" }

	-- missing function
	if not os.outputof then
		function os.outputof(cmd)
			local pipe = io.popen(cmd)
			local result = pipe:read('*a')
			pipe:close()

			result = result:gsub ("\n", " ")

			return result
		end
	end

	-- get data from shell
	LLVM_OPTIONS	= "--system-libs --libs "
	LLVM_CFLAGS	= os.outputof ("llvm-config --cflags ")
	LLVM_LFLAGS	= os.outputof ("llvm-config --ldflags " .. LLVM_OPTIONS .. "all")

	-- sanitize inputs
--	LLVM_CFLAGS	= string.gsub (LLVM_CFLAGS, "\n", " ")
--	LLVM_LFLAGS	= string.gsub (LLVM_LFLAGS, "\n", " ")
	
	-- common settings
	defines { "_GNU_SOURCE", "__STDC_LIMIT_MACROS", "__STDC_CONSTANT_MACROS" }
	buildoptions { LLVM_CFLAGS, "-pthread", "-xc", "-O0" }
	includedirs { "includes" }
	links { "dl" }
	linkoptions { LLVM_LFLAGS, "-pthread" }
	if os.is ("linux") then
		linkoptions { "-Wl,--as-needed -ltinfo" }
	end

	configuration "Debug"
		flags { "Symbols", "ExtraWarnings" }

	configuration "Release"
		buildoptions { "-march=native", "-O2" }

	-- compiler
	project "alloyc"
		kind "ConsoleApp"
		language "C++"			-- because LLVM is a bitchling
		files { "src/*.c", "src/*.h" }
