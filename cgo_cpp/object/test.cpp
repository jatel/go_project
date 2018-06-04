#include <iostream>
#include "object.hpp"
#include "test.h"

#ifdef __cplusplus
extern "C" {
#endif

void wht_print()
{
    wht_test test0;
    test0.print();

    std::string name, email;
    std::cin >> name >> email;

    wht_test test1(name);
    test1.print();

    wht_test test2(name, email);
    test2.print();
}

#ifdef __cplusplus    
}
#endif