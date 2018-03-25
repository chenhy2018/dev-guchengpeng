include(ExternalProject)
ExternalProject_Add(pjsip
        SOURCE_DIR pjproject-2.7.2
        BUILD_COMMAND make
	CONFIGURE_COMMAND "" 
	INSTALL_COMMAND ""
        BUILD_IN_SOURCE 1 

)
