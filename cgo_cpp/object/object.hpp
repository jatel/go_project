#include <string>

class wht_test
{
    public:
    wht_test(std::string name = "wanghengtao", std::string email = "jatelmomo@hotmail.com"): m_name(name), m_email(email) {}
    void print();
    private:
    std::string m_name;
    std::string m_email;
};
