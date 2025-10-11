import unittest
import os
from string_replacer import replace_strings

class TestStringReplacer(unittest.TestCase):
    def setUp(self):
        self.test_file = "test_file.txt"
        self.output_file = "output_file.txt"
        with open(self.test_file, "w") as f:
            f.write("hello world, this is a test. hello again.")

    def tearDown(self):
        if os.path.exists(self.test_file):
            os.remove(self.test_file)
        if os.path.exists(self.output_file):
            os.remove(self.output_file)

    def test_replace_strings_inplace(self):
        replace_strings(self.test_file, ["goodbye:hello,test"])
        with open(self.test_file, "r") as f:
            content = f.read()
        self.assertEqual(content, "goodbye world, this is a goodbye. goodbye again.")

    def test_multiple_substitutions_inplace(self):
        with open(self.test_file, "w") as f:
            f.write("apple banana orange apple")
        replace_strings(self.test_file, ["fruit:apple", "yellow:banana"])
        with open(self.test_file, "r") as f:
            content = f.read()
        self.assertEqual(content, "fruit yellow orange fruit")

    def test_replace_strings_to_output_file(self):
        original_content = "apple banana orange apple"
        with open(self.test_file, "w") as f:
            f.write(original_content)
        replace_strings(self.test_file, ["fruit:apple"], self.output_file)
        with open(self.output_file, "r") as f:
            output_content = f.read()
        self.assertEqual(output_content, "fruit banana orange fruit")
        with open(self.test_file, "r") as f:
            input_content = f.read()
        self.assertEqual(input_content, original_content)

if __name__ == "__main__":
    unittest.main()
