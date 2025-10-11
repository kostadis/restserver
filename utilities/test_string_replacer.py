import unittest
import os
from string_replacer import replace_strings

class TestStringReplacer(unittest.TestCase):
    def setUp(self):
        self.test_file = "test_file.txt"
        with open(self.test_file, "w") as f:
            f.write("hello world, this is a test. hello again.")

    def tearDown(self):
        os.remove(self.test_file)

    def test_replace_strings(self):
        replace_strings(self.test_file, ["goodbye:hello,test"])
        with open(self.test_file, "r") as f:
            content = f.read()
        self.assertEqual(content, "goodbye world, this is a goodbye. goodbye again.")

    def test_multiple_substitutions(self):
        with open(self.test_file, "w") as f:
            f.write("apple banana orange apple")
        replace_strings(self.test_file, ["fruit:apple", "yellow:banana"])
        with open(self.test_file, "r") as f:
            content = f.read()
        self.assertEqual(content, "fruit yellow orange fruit")

if __name__ == "__main__":
    unittest.main()
